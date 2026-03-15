package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/integration/efuncard"
	"gpt-team-api/internal/integration/meiguodizhi"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"

	"gorm.io/gorm"
)

type EfuncardClient interface {
	Redeem(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error)
	QueryCard(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error)
	Billing(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error)
	ThreeDS(ctx context.Context, code string, minutes int) (efuncard.APIResponse[efuncard.ThreeDSData], error)
}

type ProfileClient interface {
	FetchProfile(ctx context.Context) (meiguodizhi.ProfileResponse, error)
}

type CardService struct {
	cards    *repository.CardRepository
	events   *repository.CardEventRepository
	efuncard EfuncardClient
	profiles ProfileClient
}

func NewCardService(cards *repository.CardRepository, events *repository.CardEventRepository, efuncardClient EfuncardClient, profileClient ProfileClient) *CardService {
	return &CardService{
		cards:    cards,
		events:   events,
		efuncard: efuncardClient,
		profiles: profileClient,
	}
}

func (s *CardService) Import(ctx context.Context, input ImportCardsInput) (ImportResult, error) {
	if err := validateImportCardsInput(input); err != nil {
		return ImportResult{}, err
	}

	lines := strings.Split(strings.ReplaceAll(input.RawText, "\r\n", "\n"), "\n")
	seen := make(map[string]struct{}, len(lines))
	codes := make([]string, 0, len(lines))
	duplicates := make([]string, 0)

	for _, line := range lines {
		code := strings.TrimSpace(line)
		if code == "" {
			continue
		}

		if _, exists := seen[code]; exists {
			if !slices.Contains(duplicates, code) {
				duplicates = append(duplicates, code)
			}
			continue
		}

		seen[code] = struct{}{}
		codes = append(codes, code)
	}

	existing, err := s.cards.FindExistingCodes(ctx, codes)
	if err != nil {
		return ImportResult{}, apperr.Internal("card_lookup_failed", "failed to check existing cards", err)
	}

	existingSet := make(map[string]struct{}, len(existing))
	for _, code := range existing {
		existingSet[code] = struct{}{}
		if !slices.Contains(duplicates, code) {
			duplicates = append(duplicates, code)
		}
	}

	records := make([]model.Card, 0, len(codes))
	for _, code := range codes {
		if _, exists := existingSet[code]; exists {
			continue
		}

		records = append(records, model.Card{
			Code:      code,
			CardType:  input.CardType,
			CardLimit: input.CardLimit,
			Status:    model.CardStatusUnactivated,
		})
	}

	if err := s.cards.CreateMany(ctx, records); err != nil {
		return ImportResult{}, apperr.Internal("card_create_failed", "failed to create cards", err)
	}

	items := make([]CardRecord, 0, len(records))
	for _, card := range records {
		items = append(items, toCardRecord(card))
	}

	return ImportResult{
		CreatedCount: len(records),
		Duplicates:   duplicates,
		Items:        items,
	}, nil
}

func validateImportCardsInput(input ImportCardsInput) error {
	if strings.TrimSpace(input.RawText) == "" {
		return apperr.BadRequest("card_raw_text_required", "rawText is required")
	}

	switch input.CardType {
	case model.CardTypeUS, model.CardTypeUK, model.CardTypeES:
	default:
		return apperr.BadRequest("invalid_card_type", "cardType must be us, uk, or es")
	}

	switch input.CardLimit {
	case 0, 1, 2:
	default:
		return apperr.BadRequest("invalid_card_limit", "cardLimit must be 0, 1, or 2")
	}

	return nil
}

func (s *CardService) List(ctx context.Context) ([]CardRecord, error) {
	cards, err := s.cards.List(ctx)
	if err != nil {
		return nil, apperr.Internal("card_list_failed", "failed to list cards", err)
	}

	items := make([]CardRecord, 0, len(cards))
	for _, card := range cards {
		items = append(items, toCardRecord(card))
	}
	return items, nil
}

func (s *CardService) Detail(ctx context.Context, id uint64) (CardDetail, error) {
	card, err := s.findCard(ctx, id)
	if err != nil {
		return CardDetail{}, err
	}

	latestEvents, err := s.events.LatestByCard(ctx, card.ID)
	if err != nil {
		return CardDetail{}, apperr.Internal("card_event_list_failed", "failed to load card events", err)
	}

	return CardDetail{
		Card:             toCardRecord(card),
		LatestActivation: toCardEventView(latestEvents[model.CardEventActivate]),
		LatestQuery:      toCardEventView(latestEvents[model.CardEventQuery]),
		LatestBilling:    toCardEventView(latestEvents[model.CardEventBilling]),
		LatestThreeDS:    toCardEventView(latestEvents[model.CardEventThreeDS]),
		LatestIdentity:   toCardEventView(latestEvents[model.CardEventIdentityRefresh]),
	}, nil
}

func (s *CardService) Activate(ctx context.Context, id uint64) (*CardEventView, error) {
	card, err := s.findCard(ctx, id)
	if err != nil {
		return nil, err
	}

	request := map[string]any{"code": card.Code}
	response, callErr := s.efuncard.Redeem(ctx, card.Code)
	if callErr != nil {
		return s.recordFailure(ctx, card.ID, model.CardEventActivate, request, callErr)
	}

	now := time.Now().UTC()
	expiryDate := normalizeExpiryDate(
		response.Data.ExpiryDate,
		response.Data.ExpiryMonth,
		response.Data.ExpiryYear,
	)
	card.Status = model.CardStatusActivated
	card.RemoteStatus = response.Data.Status
	card.RemoteCardID = &response.Data.CardID
	card.LastFour = lastFour(response.Data.CardNumber)
	card.ExpiryDate = expiryDate
	card.LastSyncedAt = &now

	if err := s.cards.Save(ctx, &card); err != nil {
		return nil, apperr.Internal("card_save_failed", "failed to save activated card", err)
	}

	safePayload := map[string]any{
		"success": response.Success,
		"data": sanitizeCardPayload(
			response.Data.CardID,
			response.Data.Code,
			response.Data.Status,
			response.Data.Balance,
			response.Data.CreatedAt,
			response.Data.CardNumber,
			expiryDate,
			response.Data.ExpiryMonth,
			response.Data.ExpiryYear,
			response.Data.CVV,
		),
	}

	return s.recordSuccess(ctx, card.ID, model.CardEventActivate, request, safePayload, safePayload)
}

func (s *CardService) Query(ctx context.Context, id uint64) (*CardEventView, error) {
	card, err := s.findCard(ctx, id)
	if err != nil {
		return nil, err
	}

	request := map[string]any{"code": card.Code}
	response, callErr := s.efuncard.QueryCard(ctx, card.Code)
	if callErr != nil {
		return s.recordFailure(ctx, card.ID, model.CardEventQuery, request, callErr)
	}

	now := time.Now().UTC()
	expiryDate := normalizeExpiryDate(
		response.Data.ExpiryDate,
		response.Data.ExpiryMonth,
		response.Data.ExpiryYear,
	)
	card.Status = model.CardStatusActivated
	card.RemoteStatus = response.Data.Status
	card.RemoteCardID = &response.Data.CardID
	card.LastFour = lastFour(response.Data.CardNumber)
	card.ExpiryDate = expiryDate
	card.LastSyncedAt = &now

	if err := s.cards.Save(ctx, &card); err != nil {
		return nil, apperr.Internal("card_save_failed", "failed to save queried card", err)
	}

	safePayload := map[string]any{
		"success": response.Success,
		"data": sanitizeCardPayload(
			response.Data.CardID,
			response.Data.Code,
			response.Data.Status,
			response.Data.Balance,
			response.Data.CreatedAt,
			response.Data.CardNumber,
			expiryDate,
			response.Data.ExpiryMonth,
			response.Data.ExpiryYear,
			response.Data.CVV,
		),
	}
	safePayloadData, _ := safePayload["data"].(map[string]any)
	safePayloadData["balance"] = response.Data.Balance

	return s.recordSuccess(ctx, card.ID, model.CardEventQuery, request, safePayload, safePayload)
}

func (s *CardService) Billing(ctx context.Context, id uint64) (*CardEventView, error) {
	card, err := s.findCard(ctx, id)
	if err != nil {
		return nil, err
	}

	request := map[string]any{"code": card.Code}
	response, callErr := s.efuncard.Billing(ctx, card.Code)
	if callErr != nil {
		return s.recordFailure(ctx, card.ID, model.CardEventBilling, request, callErr)
	}

	safePayload := map[string]any{
		"success": response.Success,
		"data":    normalizeBillingPayload(response.Data),
	}

	return s.recordSuccess(ctx, card.ID, model.CardEventBilling, request, safePayload, safePayload)
}

func (s *CardService) ThreeDS(ctx context.Context, id uint64, minutes int) (*CardEventView, error) {
	if minutes <= 0 {
		return nil, apperr.BadRequest("invalid_minutes", "minutes must be greater than zero")
	}

	card, err := s.findCard(ctx, id)
	if err != nil {
		return nil, err
	}

	request := map[string]any{"code": card.Code, "minutes": minutes}
	response, callErr := s.efuncard.ThreeDS(ctx, card.Code, minutes)
	if callErr != nil {
		return s.recordFailure(ctx, card.ID, model.CardEventThreeDS, request, callErr)
	}

	return s.recordSuccess(ctx, card.ID, model.CardEventThreeDS, request, response, response)
}

func (s *CardService) RefreshProfile(ctx context.Context, id uint64) (*CardEventView, error) {
	card, err := s.findCard(ctx, id)
	if err != nil {
		return nil, err
	}

	request := map[string]any{"source": "meiguodizhi"}
	response, callErr := s.profiles.FetchProfile(ctx)
	if callErr != nil {
		return s.recordFailure(ctx, card.ID, model.CardEventIdentityRefresh, request, callErr)
	}

	now := time.Now().UTC()
	card.FullName = response.FullName
	card.Birthday = response.Birthday
	card.LastSyncedAt = &now

	if err := s.cards.Save(ctx, &card); err != nil {
		return nil, apperr.Internal("card_profile_save_failed", "failed to save card profile", err)
	}

	safePayload := map[string]any{
		"fullName": response.FullName,
		"birthday": response.Birthday,
	}

	return s.recordSuccess(ctx, card.ID, model.CardEventIdentityRefresh, request, response.Raw, safePayload)
}

func (s *CardService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.findCard(ctx, id); err != nil {
		return err
	}

	if err := s.cards.Delete(ctx, id); err != nil {
		return apperr.Internal("card_delete_failed", "failed to delete card", err)
	}

	return nil
}

func (s *CardService) findCard(ctx context.Context, id uint64) (model.Card, error) {
	card, err := s.cards.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Card{}, apperr.NotFound("card_not_found", "card not found")
		}
		return model.Card{}, apperr.Internal("card_lookup_failed", "failed to load card", err)
	}

	return card, nil
}

func (s *CardService) recordSuccess(ctx context.Context, cardID uint64, eventType model.CardEventType, requestBody any, responseBody any, normalized any) (*CardEventView, error) {
	event := model.CardEvent{
		CardID:            cardID,
		EventType:         eventType,
		RequestPayload:    mustJSON(requestBody),
		ResponsePayload:   mustJSON(responseBody),
		NormalizedPayload: mustJSON(normalized),
		Success:           true,
	}

	if err := s.events.Create(ctx, &event); err != nil {
		return nil, apperr.Internal("card_event_create_failed", "failed to save card event", err)
	}

	return toCardEventView(event), nil
}

func (s *CardService) recordFailure(ctx context.Context, cardID uint64, eventType model.CardEventType, requestBody any, callErr error) (*CardEventView, error) {
	event := model.CardEvent{
		CardID:            cardID,
		EventType:         eventType,
		RequestPayload:    mustJSON(requestBody),
		ResponsePayload:   "{}",
		NormalizedPayload: "{}",
		Success:           false,
		ErrorMessage:      apperr.Message(callErr),
	}

	if err := s.events.Create(ctx, &event); err != nil {
		return nil, apperr.Internal("card_event_create_failed", "failed to save failed card event", err)
	}

	return nil, callErr
}

func mustJSON(value any) string {
	if value == nil {
		return "{}"
	}

	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return "{}"
		}
		return typed
	}

	body, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}

	return string(body)
}

func lastFour(cardNumber string) string {
	if len(cardNumber) <= 4 {
		return cardNumber
	}
	return cardNumber[len(cardNumber)-4:]
}

func sanitizeCardPayload(
	cardID uint64,
	code, status string,
	balance float64,
	createdAt, cardNumber, expiryDate string,
	expiryMonth, expiryYear int,
	cvv string,
) map[string]any {
	payload := map[string]any{
		"cardId":     cardID,
		"code":       code,
		"status":     status,
		"cardNumber": cardNumber,
		"lastFour":   lastFour(cardNumber),
		"expiryDate": expiryDate,
	}
	if cvv != "" {
		payload["cvv"] = cvv
	}

	if balance != 0 {
		payload["balance"] = balance
	}
	if createdAt != "" {
		payload["createdAt"] = createdAt
	}
	if expiryMonth > 0 {
		payload["expiryMonth"] = expiryMonth
	}
	if expiryYear > 0 {
		payload["expiryYear"] = expiryYear
	}

	return payload
}

func normalizeExpiryDate(expiryDate string, expiryMonth, expiryYear int) string {
	if strings.TrimSpace(expiryDate) != "" {
		return strings.TrimSpace(expiryDate)
	}

	if expiryMonth < 1 || expiryMonth > 12 || expiryYear <= 0 {
		return ""
	}

	return fmt.Sprintf("%02d/%02d", expiryMonth, expiryYear%100)
}

func normalizeBillingPayload(data efuncard.BillingData) map[string]any {
	transactions := make([]map[string]any, 0, len(data.Transactions))
	for _, transaction := range data.Transactions {
		merchant := strings.TrimSpace(transaction.Merchant)
		if merchant == "" {
			merchant = strings.TrimSpace(transaction.MerchantName)
		}

		createdAt := strings.TrimSpace(transaction.CreatedAt)
		if createdAt == "" {
			createdAt = strings.TrimSpace(transaction.Date)
		}

		item := map[string]any{
			"id":        transaction.ID,
			"amount":    transaction.Amount,
			"currency":  transaction.Currency,
			"merchant":  merchant,
			"status":    transaction.Status,
			"createdAt": createdAt,
		}
		if transaction.MerchantName != "" {
			item["merchantName"] = transaction.MerchantName
		}
		if transaction.MerchantCity != "" {
			item["merchantCity"] = transaction.MerchantCity
		}
		if transaction.MerchantCountry != "" {
			item["merchantCountry"] = transaction.MerchantCountry
		}

		transactions = append(transactions, item)
	}

	payload := map[string]any{
		"cardId":       data.CardID,
		"code":         data.Code,
		"lastFour":     data.LastFour,
		"total":        data.Total,
		"transactions": transactions,
		"totalSpent":   data.TotalSpent,
	}

	if data.TotalSpent == 0 && data.SettledAmount != 0 {
		payload["totalSpent"] = data.SettledAmount
	}
	if data.RemainingBalance != 0 {
		payload["remainingBalance"] = data.RemainingBalance
	}
	if data.SettledCount != 0 {
		payload["settledCount"] = data.SettledCount
	}
	if data.SettledAmount != 0 {
		payload["settledAmount"] = data.SettledAmount
	}

	return payload
}

func toCardRecord(card model.Card) CardRecord {
	return CardRecord{
		ID:           card.ID,
		Code:         card.Code,
		CardType:     card.CardType,
		CardLimit:    card.CardLimit,
		Status:       card.Status,
		RemoteStatus: card.RemoteStatus,
		RemoteCardID: card.RemoteCardID,
		LastFour:     card.LastFour,
		ExpiryDate:   card.ExpiryDate,
		FullName:     card.FullName,
		Birthday:     card.Birthday,
		LastSyncedAt: card.LastSyncedAt,
		CreatedAt:    card.CreatedAt,
		UpdatedAt:    card.UpdatedAt,
	}
}

func toCardEventView(event model.CardEvent) *CardEventView {
	if event.ID == 0 {
		return nil
	}

	var payload any
	if strings.TrimSpace(event.NormalizedPayload) != "" && event.NormalizedPayload != "{}" {
		_ = json.Unmarshal([]byte(event.NormalizedPayload), &payload)
	}

	return &CardEventView{
		ID:           event.ID,
		Type:         event.EventType,
		Success:      event.Success,
		ErrorMessage: event.ErrorMessage,
		CreatedAt:    event.CreatedAt,
		Data:         payload,
	}
}

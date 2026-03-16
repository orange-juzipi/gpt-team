package service

import (
	"context"
	"strings"
	"sync"
	"testing"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/integration/efuncard"
	"gpt-team-api/internal/integration/meiguodizhi"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeEfuncardClient struct {
	redeemFn func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error)
	queryFn  func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error)
	billFn   func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error)
	threeFn  func(ctx context.Context, code string, minutes int) (efuncard.APIResponse[efuncard.ThreeDSData], error)
}

func (f fakeEfuncardClient) Redeem(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error) {
	return f.redeemFn(ctx, code)
}

func (f fakeEfuncardClient) QueryCard(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
	return f.queryFn(ctx, code)
}

func (f fakeEfuncardClient) Billing(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error) {
	return f.billFn(ctx, code)
}

func (f fakeEfuncardClient) ThreeDS(ctx context.Context, code string, minutes int) (efuncard.APIResponse[efuncard.ThreeDSData], error) {
	return f.threeFn(ctx, code, minutes)
}

type fakeProfileClient struct {
	fetchFn func(ctx context.Context, cardType model.CardType) (meiguodizhi.ProfileResponse, error)
}

func (f fakeProfileClient) FetchProfile(ctx context.Context, cardType model.CardType) (meiguodizhi.ProfileResponse, error) {
	return f.fetchFn(ctx, cardType)
}

func TestImportCardsDeduplicatesAndSkipsExisting(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	service := NewCardService(cardRepo, eventRepo, fakeEfuncardClient{}, fakeProfileClient{})

	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "EXISTING", CardType: model.CardTypeUS, CardLimit: 0, Status: model.CardStatusUnactivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	result, err := service.Import(context.Background(), ImportCardsInput{
		RawText:   "ONE\nEXISTING\nONE\nTWO\n",
		CardType:  model.CardTypeUK,
		CardLimit: 2,
	})
	if err != nil {
		t.Fatalf("import cards: %v", err)
	}

	if result.CreatedCount != 2 {
		t.Fatalf("expected 2 cards created, got %d", result.CreatedCount)
	}

	if len(result.Duplicates) != 2 {
		t.Fatalf("expected duplicate list size 2, got %d", len(result.Duplicates))
	}

	if result.Items[0].CardType != model.CardTypeUK || result.Items[0].CardLimit != 2 {
		t.Fatalf("expected imported cards to keep chosen type and limit, got %+v", result.Items[0])
	}
}

func TestActivateCardStoresSnapshotForDetailView(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)

	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-1", CardType: model.CardTypeUS, CardLimit: 1, Status: model.CardStatusUnactivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{
			redeemFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error) {
				return efuncard.APIResponse[efuncard.RedeemData]{
					Success: true,
					Data: efuncard.RedeemData{
						CardID:     10,
						CardNumber: "4111111111111111",
						CVV:        "123",
						ExpiryDate: "12/25",
						Code:       code,
						Status:     "active",
					},
				}, nil
			},
			queryFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
				return efuncard.APIResponse[efuncard.QueryData]{}, nil
			},
			billFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error) {
				return efuncard.APIResponse[efuncard.BillingData]{}, nil
			},
			threeFn: func(ctx context.Context, code string, minutes int) (efuncard.APIResponse[efuncard.ThreeDSData], error) {
				return efuncard.APIResponse[efuncard.ThreeDSData]{}, nil
			},
		},
		fakeProfileClient{},
	)

	event, err := service.Activate(context.Background(), 1)
	if err != nil {
		t.Fatalf("activate card: %v", err)
	}

	if event == nil || !event.Success {
		t.Fatalf("expected success event")
	}

	card, err := cardRepo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("reload card: %v", err)
	}

	if card.Status != model.CardStatusActivated {
		t.Fatalf("expected activated status, got %s", card.Status)
	}

	if card.LastFour != "1111" {
		t.Fatalf("expected last four 1111, got %s", card.LastFour)
	}

	latest, err := eventRepo.LatestByCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("latest events: %v", err)
	}

	payload := latest[model.CardEventActivate].NormalizedPayload
	if !containsAll(payload, "4111111111111111", "\"cvv\":\"123\"", "\"lastFour\":\"1111\"") {
		t.Fatalf("expected card snapshot payload to include card number and cvv, got: %s", payload)
	}
}

func TestActivateAlreadyUsedSyncsCardStatusViaQuery(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)

	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-USED", CardType: model.CardTypeUK, CardLimit: 0, Status: model.CardStatusUnactivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{
			redeemFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error) {
				return efuncard.APIResponse[efuncard.RedeemData]{}, apperr.Conflict("efuncard_http_409", "激活码已使用")
			},
			queryFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
				return efuncard.APIResponse[efuncard.QueryData]{
					Success: true,
					Data: efuncard.QueryData{
						CardID:      66,
						CardNumber:  "4242424242424242",
						CVV:         "456",
						ExpiryMonth: 9,
						ExpiryYear:  2029,
						Code:        code,
						Status:      "ACTIVE",
						Balance:     0,
					},
				}, nil
			},
			billFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error) {
				return efuncard.APIResponse[efuncard.BillingData]{}, nil
			},
			threeFn: func(ctx context.Context, code string, minutes int) (efuncard.APIResponse[efuncard.ThreeDSData], error) {
				return efuncard.APIResponse[efuncard.ThreeDSData]{}, nil
			},
		},
		fakeProfileClient{},
	)

	event, err := service.Activate(context.Background(), 1)
	if err != nil {
		t.Fatalf("activate card with sync fallback: %v", err)
	}

	if event == nil || !event.Success {
		t.Fatalf("expected activation event success after query fallback")
	}

	card, err := cardRepo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("reload card: %v", err)
	}

	if card.Status != model.CardStatusActivated {
		t.Fatalf("expected fallback query to promote card status, got %s", card.Status)
	}

	if card.LastFour != "4242" {
		t.Fatalf("expected last four 4242, got %s", card.LastFour)
	}

	latest, err := eventRepo.LatestByCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("latest events: %v", err)
	}

	payload := latest[model.CardEventActivate].NormalizedPayload
	if !containsAll(payload, `"autoSynced":true`, `"message":"激活码已使用，已自动同步卡片状态"`, `"lastFour":"4242"`) {
		t.Fatalf("expected normalized payload to include fallback sync details, got: %s", payload)
	}
}

func TestQueryMarksCardActivated(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-QUERY", CardType: model.CardTypeES, CardLimit: 2, Status: model.CardStatusUnactivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{
			redeemFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error) {
				return efuncard.APIResponse[efuncard.RedeemData]{}, nil
			},
			queryFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
				return efuncard.APIResponse[efuncard.QueryData]{
					Success: true,
					Data: efuncard.QueryData{
						CardID:     22,
						CardNumber: "5555555555552222",
						ExpiryDate: "08/29",
						Code:       code,
						Status:     "active",
						Balance:    90,
					},
				}, nil
			},
			billFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error) {
				return efuncard.APIResponse[efuncard.BillingData]{}, nil
			},
			threeFn: func(ctx context.Context, code string, minutes int) (efuncard.APIResponse[efuncard.ThreeDSData], error) {
				return efuncard.APIResponse[efuncard.ThreeDSData]{}, nil
			},
		},
		fakeProfileClient{},
	)

	if _, err := service.Query(context.Background(), 1); err != nil {
		t.Fatalf("query card: %v", err)
	}

	card, err := cardRepo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("reload card: %v", err)
	}

	if card.Status != model.CardStatusActivated {
		t.Fatalf("expected query to promote card status to activated, got %s", card.Status)
	}

	latest, err := eventRepo.LatestByCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("latest events: %v", err)
	}

	payload := latest[model.CardEventQuery].NormalizedPayload
	if !strings.Contains(payload, `"balance":90`) {
		t.Fatalf("expected query snapshot payload to include balance, got: %s", payload)
	}
}

func TestQueryStoresZeroBalance(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-ZERO", CardType: model.CardTypeUK, CardLimit: 0, Status: model.CardStatusUnactivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{
			redeemFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error) {
				return efuncard.APIResponse[efuncard.RedeemData]{}, nil
			},
			queryFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
				return efuncard.APIResponse[efuncard.QueryData]{
					Success: true,
					Data: efuncard.QueryData{
						CardID:     30,
						CardNumber: "4000000000000002",
						ExpiryDate: "",
						Code:       code,
						Status:     "ACTIVE",
						Balance:    0,
					},
				}, nil
			},
			billFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error) {
				return efuncard.APIResponse[efuncard.BillingData]{}, nil
			},
			threeFn: func(ctx context.Context, code string, minutes int) (efuncard.APIResponse[efuncard.ThreeDSData], error) {
				return efuncard.APIResponse[efuncard.ThreeDSData]{}, nil
			},
		},
		fakeProfileClient{},
	)

	if _, err := service.Query(context.Background(), 1); err != nil {
		t.Fatalf("query card: %v", err)
	}

	latest, err := eventRepo.LatestByCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("latest events: %v", err)
	}

	payload := latest[model.CardEventQuery].NormalizedPayload
	if !strings.Contains(payload, `"balance":0`) {
		t.Fatalf("expected zero balance to be kept in snapshot payload, got: %s", payload)
	}
}

func TestQueryBuildsExpiryDateFromMonthYear(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-EXPIRY", CardType: model.CardTypeUK, CardLimit: 0, Status: model.CardStatusUnactivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{
			queryFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
				return efuncard.APIResponse[efuncard.QueryData]{
					Success: true,
					Data: efuncard.QueryData{
						CardID:      24119,
						CardNumber:  "4462220001292405",
						ExpiryMonth: 3,
						ExpiryYear:  2029,
						CVV:         "421",
						Code:        code,
						Status:      "CANCELLED",
					},
				}, nil
			},
		},
		fakeProfileClient{},
	)

	if _, err := service.Query(context.Background(), 1); err != nil {
		t.Fatalf("query card: %v", err)
	}

	card, err := cardRepo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("reload card: %v", err)
	}

	if card.ExpiryDate != "03/29" {
		t.Fatalf("expected derived expiry date 03/29, got %q", card.ExpiryDate)
	}

	latest, err := eventRepo.LatestByCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("latest events: %v", err)
	}

	payload := latest[model.CardEventQuery].NormalizedPayload
	if !strings.Contains(payload, `"expiryMonth":3`) || !strings.Contains(payload, `"expiryYear":2029`) {
		t.Fatalf("expected query snapshot payload to keep expiry month/year, got: %s", payload)
	}
}

func TestQueryPrefersMonthYearAndKeepsPreciseExpiryTime(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-EXPIRES-AT", CardType: model.CardTypeUK, CardLimit: 0, Status: model.CardStatusUnactivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{
			queryFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
				return efuncard.APIResponse[efuncard.QueryData]{
					Success: true,
					Data: efuncard.QueryData{
						CardID:      24119,
						CardNumber:  "4462220001292405",
						ExpiryDate:  "03/01",
						ExpiryMonth: 3,
						ExpiryYear:  2029,
						ExpiresAt:   "2026-03-16T02:07:48Z",
						CVV:         "421",
						Code:        code,
						Status:      "ACTIVE",
					},
				}, nil
			},
		},
		fakeProfileClient{},
	)

	if _, err := service.Query(context.Background(), 1); err != nil {
		t.Fatalf("query card: %v", err)
	}

	card, err := cardRepo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("reload card: %v", err)
	}

	if card.ExpiryDate != "03/29" {
		t.Fatalf("expected normalized expiry date 03/29, got %q", card.ExpiryDate)
	}

	latest, err := eventRepo.LatestByCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("latest events: %v", err)
	}

	payload := latest[model.CardEventQuery].NormalizedPayload
	if !strings.Contains(payload, `"expiresAt":"2026-03-16T02:07:48Z"`) {
		t.Fatalf("expected query snapshot payload to keep expiresAt, got: %s", payload)
	}
}

func TestQueryKeepsBillingAddressInstructions(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-ADDRESS", CardType: model.CardTypeUK, CardLimit: 0, Status: model.CardStatusActivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{
			queryFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
				return efuncard.APIResponse[efuncard.QueryData]{
					Success: true,
					Data: efuncard.QueryData{
						CardID:            24119,
						CardNumber:        "4462220001292405",
						ExpiryMonth:       3,
						ExpiryYear:        2029,
						CVV:               "421",
						Code:              code,
						Status:            "ACTIVE",
						NodeInstructions:  "123 Main St<br>Los Angeles, CA 90001",
						GroupInstructions: "Use US billing details only",
					},
				}, nil
			},
		},
		fakeProfileClient{},
	)

	if _, err := service.Query(context.Background(), 1); err != nil {
		t.Fatalf("query card: %v", err)
	}

	latest, err := eventRepo.LatestByCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("latest events: %v", err)
	}

	payload := latest[model.CardEventQuery].NormalizedPayload
	if !strings.Contains(payload, `"billingAddress":"123 Main St\u003cbr\u003eLos Angeles, CA 90001"`) {
		t.Fatalf("expected query snapshot payload to keep billing address, got: %s", payload)
	}
	if !strings.Contains(payload, `"groupInstructions":"Use US billing details only"`) {
		t.Fatalf("expected query snapshot payload to keep instructions, got: %s", payload)
	}
}

func TestBillingNormalizesMerchantNameAndDate(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-BILLING", CardType: model.CardTypeUK, CardLimit: 0, Status: model.CardStatusActivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{
			billFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error) {
				return efuncard.APIResponse[efuncard.BillingData]{
					Success: true,
					Data: efuncard.BillingData{
						CardID:           24119,
						LastFour:         "2405",
						Total:            1,
						SettledCount:     0,
						SettledAmount:    0,
						NodeInstructions: "123 Main St<br>Los Angeles, CA 90001",
						Transactions: []efuncard.BillingTransaction{
							{
								ID:           "txn-1",
								Amount:       0,
								Currency:     "USD",
								MerchantName: "OPENAI",
								Status:       "APPROVED",
								Date:         "2026-03-15T09:25:06.428+0000",
							},
						},
					},
				}, nil
			},
		},
		fakeProfileClient{},
	)

	if _, err := service.Billing(context.Background(), 1); err != nil {
		t.Fatalf("billing card: %v", err)
	}

	latest, err := eventRepo.LatestByCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("latest events: %v", err)
	}

	payload := latest[model.CardEventBilling].NormalizedPayload
	if !strings.Contains(payload, `"merchant":"OPENAI"`) {
		t.Fatalf("expected merchant to be normalized, got: %s", payload)
	}
	if !strings.Contains(payload, `"createdAt":"2026-03-15T09:25:06.428+0000"`) {
		t.Fatalf("expected createdAt to be normalized, got: %s", payload)
	}
	if !strings.Contains(payload, `"billingAddress":"123 Main St\u003cbr\u003eLos Angeles, CA 90001"`) {
		t.Fatalf("expected billing payload to keep billing address, got: %s", payload)
	}
}

func TestRefreshProfileUpdatesCard(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-PROFILE", CardType: model.CardTypeUK, CardLimit: 0, Status: model.CardStatusActivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{},
		fakeProfileClient{
			fetchFn: func(ctx context.Context, cardType model.CardType) (meiguodizhi.ProfileResponse, error) {
				if cardType != model.CardTypeUK {
					t.Fatalf("expected UK card type, got %s", cardType)
				}
				return meiguodizhi.ProfileResponse{
					FullName:      "Ada Lovelace",
					Birthday:      "1815-12-10",
					StreetAddress: "2350 Monroe Street",
					City:          "Houston",
					State:         "TX",
					StateFull:     "Texas",
					ZipCode:       "77028",
					PhoneNumber:   "713-375-5326",
					Raw:           `{"fullName":"Ada Lovelace","birthday":"1815-12-10"}`,
				}, nil
			},
		},
	)

	if _, err := service.RefreshProfile(context.Background(), 1); err != nil {
		t.Fatalf("refresh profile: %v", err)
	}

	card, err := cardRepo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("reload card: %v", err)
	}

	if card.FullName != "Ada Lovelace" || card.Birthday != "1815-12-10" {
		t.Fatalf("profile not saved, got %q / %q", card.FullName, card.Birthday)
	}
	if card.StreetAddress != "2350 Monroe Street" || card.ZipCode != "77028" || card.PhoneNumber != "713-375-5326" {
		t.Fatalf("address fields not saved, got %+v", card)
	}
}

func TestRefreshProfileKeepsLatestResultDuringConcurrentRequests(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	cardRepo := repository.NewCardRepository(db)
	eventRepo := repository.NewCardEventRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-PROFILE-CONCURRENT", CardType: model.CardTypeUK, CardLimit: 0, Status: model.CardStatusActivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var once sync.Once
	var callCount int
	var callMu sync.Mutex

	service := NewCardService(
		cardRepo,
		eventRepo,
		fakeEfuncardClient{},
		fakeProfileClient{
			fetchFn: func(ctx context.Context, cardType model.CardType) (meiguodizhi.ProfileResponse, error) {
				callMu.Lock()
				callCount++
				currentCall := callCount
				callMu.Unlock()

				if currentCall == 1 {
					once.Do(func() {
						close(firstStarted)
					})
					<-releaseFirst
					return meiguodizhi.ProfileResponse{
						FullName: "Old Profile",
						Birthday: "1999-01-01",
						Raw:      `{"fullName":"Old Profile","birthday":"1999-01-01"}`,
					}, nil
				}

				return meiguodizhi.ProfileResponse{
					FullName: "New Profile",
					Birthday: "2001-10-01",
					Raw:      `{"fullName":"New Profile","birthday":"2001-10-01"}`,
				}, nil
			},
		},
	)

	firstErrCh := make(chan error, 1)
	go func() {
		_, err := service.RefreshProfile(context.Background(), 1)
		firstErrCh <- err
	}()

	<-firstStarted

	secondErrCh := make(chan error, 1)
	go func() {
		_, err := service.RefreshProfile(context.Background(), 1)
		secondErrCh <- err
	}()

	close(releaseFirst)

	if err := <-firstErrCh; err != nil {
		t.Fatalf("first refresh profile: %v", err)
	}
	if err := <-secondErrCh; err != nil {
		t.Fatalf("second refresh profile: %v", err)
	}

	card, err := cardRepo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("reload card: %v", err)
	}

	if card.FullName != "New Profile" || card.Birthday != "2001-10-01" {
		t.Fatalf("expected latest profile to win, got %q / %q", card.FullName, card.Birthday)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.Card{}, &model.CardEvent{}, &model.Account{}, &model.MailboxProvider{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	return db
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func containsAll(value string, needles ...string) bool {
	for _, needle := range needles {
		if needle == "" {
			continue
		}
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}

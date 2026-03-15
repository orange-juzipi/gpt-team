package service

import (
	"context"
	"errors"
	"strings"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"

	"gorm.io/gorm"
)

type MailboxProviderService struct {
	providers *repository.MailboxProviderRepository
	cipher    *Cipher
}

type MailboxProviderConfig struct {
	ProviderType model.MailboxProviderType
	AccountEmail string
	Password     string
}

func NewMailboxProviderService(providers *repository.MailboxProviderRepository, cipher *Cipher) *MailboxProviderService {
	return &MailboxProviderService{
		providers: providers,
		cipher:    cipher,
	}
}

func (s *MailboxProviderService) List(ctx context.Context) ([]MailboxProviderRecord, error) {
	items, err := s.providers.List(ctx)
	if err != nil {
		return nil, apperr.Internal("mailbox_provider_list_failed", "failed to list mailbox providers", err)
	}

	records := make([]MailboxProviderRecord, 0, len(items))
	for _, item := range items {
		record, err := s.toRecord(item)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func (s *MailboxProviderService) Create(ctx context.Context, input MailboxProviderInput) (MailboxProviderRecord, error) {
	providerType, err := normalizeMailboxProviderType(input.ProviderType)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	normalizedSuffix, err := normalizeMailboxDomainSuffix(input.DomainSuffix)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	normalizedAccountEmail, normalizedPassword, err := normalizeMailboxCredentials(providerType, input.AccountEmail, input.Password)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	if _, err := s.providers.FindByDomainSuffix(ctx, normalizedSuffix); err == nil {
		return MailboxProviderRecord{}, apperr.Conflict("mailbox_provider_suffix_exists", "domain suffix already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return MailboxProviderRecord{}, apperr.Internal("mailbox_provider_lookup_failed", "failed to inspect mailbox provider suffix", err)
	}

	encryptedPassword, err := s.cipher.Encrypt(normalizedPassword)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	item := model.MailboxProvider{
		ProviderType:    providerType,
		AccountEmail:    normalizedAccountEmail,
		DomainSuffix:    normalizedSuffix,
		AccountID:       nil,
		TokenCiphertext: encryptedPassword,
		Remark:          strings.TrimSpace(input.Remark),
	}

	if err := s.providers.Create(ctx, &item); err != nil {
		return MailboxProviderRecord{}, apperr.Internal("mailbox_provider_create_failed", "failed to create mailbox provider", err)
	}

	return s.toRecord(item)
}

func (s *MailboxProviderService) Update(ctx context.Context, id uint64, input MailboxProviderInput) (MailboxProviderRecord, error) {
	item, err := s.findProvider(ctx, id)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	providerType, err := normalizeMailboxProviderType(input.ProviderType)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	normalizedSuffix, err := normalizeMailboxDomainSuffix(input.DomainSuffix)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	normalizedAccountEmail, normalizedPassword, err := normalizeMailboxCredentials(providerType, input.AccountEmail, input.Password)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	existing, err := s.providers.FindByDomainSuffix(ctx, normalizedSuffix)
	if err == nil && existing.ID != item.ID {
		return MailboxProviderRecord{}, apperr.Conflict("mailbox_provider_suffix_exists", "domain suffix already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return MailboxProviderRecord{}, apperr.Internal("mailbox_provider_lookup_failed", "failed to inspect mailbox provider suffix", err)
	}

	encryptedPassword, err := s.cipher.Encrypt(normalizedPassword)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	item.ProviderType = providerType
	item.AccountEmail = normalizedAccountEmail
	item.DomainSuffix = normalizedSuffix
	item.AccountID = nil
	item.TokenCiphertext = encryptedPassword
	item.Remark = strings.TrimSpace(input.Remark)

	if err := s.providers.Save(ctx, &item); err != nil {
		return MailboxProviderRecord{}, apperr.Internal("mailbox_provider_update_failed", "failed to update mailbox provider", err)
	}

	return s.toRecord(item)
}

func (s *MailboxProviderService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.findProvider(ctx, id); err != nil {
		return err
	}

	if err := s.providers.Delete(ctx, id); err != nil {
		return apperr.Internal("mailbox_provider_delete_failed", "failed to delete mailbox provider", err)
	}

	return nil
}

func (s *MailboxProviderService) ResolveConfigByAccount(ctx context.Context, account string) (MailboxProviderConfig, bool, error) {
	suffix, ok := extractMailboxDomainSuffix(account)
	if !ok {
		return MailboxProviderConfig{}, false, nil
	}

	item, err := s.providers.FindByDomainSuffix(ctx, suffix)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return MailboxProviderConfig{}, false, nil
		}
		return MailboxProviderConfig{}, false, apperr.Internal("mailbox_provider_lookup_failed", "failed to inspect mailbox provider suffix", err)
	}

	password, err := s.cipher.Decrypt(item.TokenCiphertext)
	if err != nil {
		return MailboxProviderConfig{}, false, err
	}

	providerType, err := normalizeMailboxProviderType(item.ProviderType)
	if err != nil {
		return MailboxProviderConfig{}, false, err
	}

	return MailboxProviderConfig{
		ProviderType: providerType,
		AccountEmail: item.AccountEmail,
		Password:     password,
	}, true, nil
}

func (s *MailboxProviderService) findProvider(ctx context.Context, id uint64) (model.MailboxProvider, error) {
	item, err := s.providers.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.MailboxProvider{}, apperr.NotFound("mailbox_provider_not_found", "mailbox provider not found")
		}

		return model.MailboxProvider{}, apperr.Internal("mailbox_provider_lookup_failed", "failed to load mailbox provider", err)
	}

	return item, nil
}

func (s *MailboxProviderService) toRecord(item model.MailboxProvider) (MailboxProviderRecord, error) {
	password, err := s.cipher.Decrypt(item.TokenCiphertext)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	providerType, err := normalizeMailboxProviderType(item.ProviderType)
	if err != nil {
		return MailboxProviderRecord{}, err
	}

	return MailboxProviderRecord{
		ID:             item.ID,
		ProviderType:   providerType,
		DomainSuffix:   item.DomainSuffix,
		AccountEmail:   item.AccountEmail,
		Password:       password,
		MaskedPassword: maskPassword(password),
		Remark:         item.Remark,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}, nil
}

func normalizeMailboxProviderType(value model.MailboxProviderType) (model.MailboxProviderType, error) {
	switch model.MailboxProviderType(strings.ToLower(strings.TrimSpace(string(value)))) {
	case "":
		return model.MailboxProviderTypeCloudmail, nil
	case model.MailboxProviderTypeCloudmail:
		return model.MailboxProviderTypeCloudmail, nil
	case model.MailboxProviderTypeDuckmail:
		return model.MailboxProviderTypeDuckmail, nil
	default:
		return "", apperr.BadRequest("mailbox_provider_type_invalid", "providerType must be cloudmail or duckmail")
	}
}

func normalizeMailboxCredentials(providerType model.MailboxProviderType, accountEmail, password string) (string, string, error) {
	trimmedPassword := strings.TrimSpace(password)

	switch providerType {
	case model.MailboxProviderTypeDuckmail:
		if trimmedPassword == "" {
			return "", "", apperr.BadRequest("mailbox_provider_api_key_required", "DuckMail api key is required")
		}
		if !strings.HasPrefix(trimmedPassword, "dk_") {
			return "", "", apperr.BadRequest("mailbox_provider_api_key_invalid", "DuckMail api key must start with dk_")
		}
		if strings.TrimSpace(accountEmail) == "" {
			return "", trimmedPassword, nil
		}

		normalizedAccountEmail, err := normalizeMailboxAccountEmail(accountEmail)
		if err != nil {
			return "", "", err
		}
		return normalizedAccountEmail, trimmedPassword, nil
	default:
		normalizedAccountEmail, err := normalizeMailboxAccountEmail(accountEmail)
		if err != nil {
			return "", "", err
		}
		if trimmedPassword == "" {
			return "", "", apperr.BadRequest("mailbox_provider_password_required", "password is required")
		}
		return normalizedAccountEmail, trimmedPassword, nil
	}
}

func normalizeMailboxDomainSuffix(value string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	trimmed = strings.TrimPrefix(trimmed, "@")
	if trimmed == "" {
		return "", apperr.BadRequest("mailbox_provider_suffix_required", "domainSuffix is required")
	}

	if strings.Contains(trimmed, "@") {
		return "", apperr.BadRequest("mailbox_provider_suffix_invalid", "domainSuffix must only contain the email suffix")
	}

	return trimmed, nil
}

func normalizeMailboxAccountEmail(value string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "", apperr.BadRequest("mailbox_provider_account_email_required", "accountEmail is required")
	}
	if strings.Count(trimmed, "@") != 1 {
		return "", apperr.BadRequest("mailbox_provider_account_email_invalid", "accountEmail must be a valid email address")
	}

	return trimmed, nil
}

func extractMailboxDomainSuffix(account string) (string, bool) {
	trimmed := strings.ToLower(strings.TrimSpace(account))
	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/integration/mailbox"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"

	"gorm.io/gorm"
)

type AccountMailClient interface {
	ListInboxEmails(ctx context.Context, authEmail, authPassword, targetEmail string) ([]mailbox.Email, error)
}

type DuckmailClient interface {
	AccountMailClient
	CreateAccount(ctx context.Context, apiKey, address, password string) error
}

type MailboxTokenResolver interface {
	ResolveConfigByAccount(ctx context.Context, account string) (MailboxProviderConfig, bool, error)
}

type AccountService struct {
	accounts      *repository.AccountRepository
	cipher        *Cipher
	cloudmailer   AccountMailClient
	duckmailer    DuckmailClient
	mailboxTokens MailboxTokenResolver
}

func NewAccountService(accounts *repository.AccountRepository, cipher *Cipher, cloudmailer AccountMailClient, duckmailer DuckmailClient, mailboxTokens MailboxTokenResolver) *AccountService {
	return &AccountService{
		accounts:      accounts,
		cipher:        cipher,
		cloudmailer:   cloudmailer,
		duckmailer:    duckmailer,
		mailboxTokens: mailboxTokens,
	}
}

func (s *AccountService) List(ctx context.Context) ([]AccountRecord, error) {
	accounts, err := s.accounts.ListTopLevel(ctx)
	if err != nil {
		return nil, apperr.Internal("account_list_failed", "failed to list accounts", err)
	}

	return s.toRecords(accounts)
}

func (s *AccountService) Create(ctx context.Context, input AccountInput) (AccountRecord, error) {
	if err := validateAccountInput(input); err != nil {
		return AccountRecord{}, err
	}

	if input.CreateMailbox {
		if err := s.createDuckmailAccountIfNeeded(ctx, strings.TrimSpace(input.Account), input.Password); err != nil {
			return AccountRecord{}, err
		}
	}

	encrypted, err := s.cipher.Encrypt(input.Password)
	if err != nil {
		return AccountRecord{}, err
	}

	account := model.Account{
		Account:            strings.TrimSpace(input.Account),
		PasswordCiphertext: encrypted,
		Type:               input.Type,
		StartTime:          input.StartTime,
		EndTime:            input.EndTime,
		Status:             input.Status,
		Remark:             strings.TrimSpace(input.Remark),
	}

	if err := s.accounts.Create(ctx, &account); err != nil {
		return AccountRecord{}, apperr.Internal("account_create_failed", "failed to create account", err)
	}

	return s.toRecord(account)
}

func (s *AccountService) Update(ctx context.Context, id uint64, input AccountInput) (AccountRecord, error) {
	if err := validateAccountInput(input); err != nil {
		return AccountRecord{}, err
	}

	account, err := s.findAccount(ctx, id)
	if err != nil {
		return AccountRecord{}, err
	}

	if account.ParentID != nil {
		return AccountRecord{}, apperr.BadRequest("account_not_top_level", "child accounts must be updated from their dedicated endpoints")
	}

	if account.Status == model.AccountStatusBlocked && input.Status == model.AccountStatusNormal {
		childrenCount, err := s.accounts.CountChildrenByRelation(ctx, account.ID, model.AccountRelationTypeWarranty)
		if err != nil {
			return AccountRecord{}, apperr.Internal("account_child_count_failed", "failed to inspect warranty accounts", err)
		}

		if childrenCount > 0 {
			return AccountRecord{}, apperr.Conflict("account_has_warranties", "blocked account cannot return to normal while warranty accounts exist")
		}
	}

	encrypted, err := s.cipher.Encrypt(input.Password)
	if err != nil {
		return AccountRecord{}, err
	}

	account.Account = strings.TrimSpace(input.Account)
	account.PasswordCiphertext = encrypted
	account.Type = input.Type
	account.StartTime = input.StartTime
	account.EndTime = input.EndTime
	account.Status = input.Status
	account.Remark = strings.TrimSpace(input.Remark)

	if err := s.accounts.Save(ctx, &account); err != nil {
		return AccountRecord{}, apperr.Internal("account_update_failed", "failed to update account", err)
	}

	return s.toRecord(account)
}

func (s *AccountService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.findAccount(ctx, id); err != nil {
		return err
	}

	if err := s.accounts.DeleteCascade(ctx, id); err != nil {
		return apperr.Internal("account_delete_failed", "failed to delete account", err)
	}

	return nil
}

func (s *AccountService) ListWarranties(ctx context.Context, parentID uint64) ([]AccountRecord, error) {
	parent, err := s.findAccount(ctx, parentID)
	if err != nil {
		return nil, err
	}

	if err := validateWarrantyParent(parent); err != nil {
		return nil, err
	}

	accounts, err := s.accounts.ListChildrenByRelation(ctx, parentID, model.AccountRelationTypeWarranty)
	if err != nil {
		return nil, apperr.Internal("account_warranty_list_failed", "failed to list warranty accounts", err)
	}

	return s.toRecords(accounts)
}

func (s *AccountService) ListSubAccounts(ctx context.Context, parentID uint64) ([]AccountRecord, error) {
	parent, err := s.findAccount(ctx, parentID)
	if err != nil {
		return nil, err
	}

	if err := validateSubAccountParent(parent); err != nil {
		return nil, err
	}

	accounts, err := s.accounts.ListChildrenByRelation(ctx, parentID, model.AccountRelationTypeSubAccount)
	if err != nil {
		return nil, apperr.Internal("account_subaccount_list_failed", "failed to list sub accounts", err)
	}

	return s.toRecords(accounts)
}

func (s *AccountService) ListEmails(ctx context.Context, accountID uint64) (AccountEmailList, error) {
	account, err := s.findAccount(ctx, accountID)
	if err != nil {
		return AccountEmailList{}, err
	}

	if s.cloudmailer == nil {
		return AccountEmailList{}, apperr.Internal("cloudmail_client_missing", "Cloudmail client is not configured", nil)
	}

	password, err := s.cipher.Decrypt(account.PasswordCiphertext)
	if err != nil {
		return AccountEmailList{}, err
	}

	providerType := model.MailboxProviderTypeCloudmail
	authEmail := account.Account
	authPassword := password
	usesProviderCloudmailCredentials := false
	if s.mailboxTokens != nil {
		resolvedConfig, matched, err := s.mailboxTokens.ResolveConfigByAccount(ctx, account.Account)
		if err != nil {
			return AccountEmailList{}, err
		}
		if matched {
			providerType = resolvedConfig.ProviderType
			if providerType == model.MailboxProviderTypeCloudmail {
				authEmail = resolvedConfig.AccountEmail
				authPassword = resolvedConfig.Password
				usesProviderCloudmailCredentials = authEmail != "" && authEmail != account.Account
			}
		}
	}

	mailer := s.cloudmailer
	if providerType == model.MailboxProviderTypeDuckmail {
		if s.duckmailer == nil {
			return AccountEmailList{}, apperr.Internal("duckmail_client_missing", "DuckMail client is not configured", nil)
		}
		mailer = s.duckmailer
		authEmail = account.Account
		authPassword = password
	}

	items, err := mailer.ListInboxEmails(ctx, authEmail, authPassword, account.Account)
	if err != nil && shouldRetryCloudmailWithAccountCredentials(err, providerType, usesProviderCloudmailCredentials) {
		items, err = s.cloudmailer.ListInboxEmails(ctx, account.Account, password, account.Account)
	}
	if err != nil {
		return AccountEmailList{}, err
	}

	emails := make([]AccountEmailRecord, 0, len(items))
	for _, item := range items {
		emails = append(emails, toAccountEmailRecord(item))
	}

	return AccountEmailList{
		AccountID: account.ID,
		Account:   account.Account,
		Items:     emails,
	}, nil
}

func shouldRetryCloudmailWithAccountCredentials(err error, providerType model.MailboxProviderType, usesProviderCloudmailCredentials bool) bool {
	if providerType != model.MailboxProviderTypeCloudmail || !usesProviderCloudmailCredentials {
		return false
	}

	status := apperr.Status(err)
	if status != 400 && status != 401 && status != 403 {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(apperr.Message(err)))
	if message == "" {
		return false
	}

	return strings.Contains(message, "invalid email or password") ||
		strings.Contains(message, "invalid email") ||
		strings.Contains(message, "invalid password")
}

func (s *AccountService) CreateWarranty(ctx context.Context, parentID uint64, input AccountInput) (AccountRecord, error) {
	parent, err := s.findAccount(ctx, parentID)
	if err != nil {
		return AccountRecord{}, err
	}

	if err := validateWarrantyParent(parent); err != nil {
		return AccountRecord{}, err
	}

	return s.createChildAccount(ctx, parentID, input, model.AccountRelationTypeWarranty, "warranty_create_failed", "failed to create warranty account")
}

func (s *AccountService) CreateSubAccount(ctx context.Context, parentID uint64, input AccountInput) (AccountRecord, error) {
	parent, err := s.findAccount(ctx, parentID)
	if err != nil {
		return AccountRecord{}, err
	}

	if err := validateSubAccountParent(parent); err != nil {
		return AccountRecord{}, err
	}

	if input.UseServerTimeSchedule {
		now := time.Now().UTC()
		input.StartTime = &now
		input.EndTime = &now
	}

	return s.createChildAccount(ctx, parentID, input, model.AccountRelationTypeSubAccount, "subaccount_create_failed", "failed to create sub account")
}

func (s *AccountService) createDuckmailAccountIfNeeded(ctx context.Context, account, password string) error {
	if s.mailboxTokens == nil {
		return nil
	}

	config, matched, err := s.mailboxTokens.ResolveConfigByAccount(ctx, account)
	if err != nil {
		return err
	}
	if !matched || config.ProviderType != model.MailboxProviderTypeDuckmail {
		return nil
	}
	if s.duckmailer == nil {
		return apperr.Internal("duckmail_client_missing", "DuckMail client is not configured", nil)
	}

	if err := s.duckmailer.CreateAccount(ctx, config.Password, account, password); err != nil {
		message := resolveDuckmailCreateErrorMessage(err)
		status := apperr.Status(err)
		if status >= 400 && status < 500 {
			return apperr.New(status, "duckmail_account_create_failed", message)
		}

		return apperr.Upstream("duckmail_account_create_failed", message, err)
	}

	return nil
}

func resolveDuckmailCreateErrorMessage(err error) string {
	upstreamMessage := strings.TrimSpace(apperr.Message(err))
	if upstreamMessage == "" || upstreamMessage == "internal server error" {
		return "DuckMail 邮箱创建失败，请检查邮箱管理中的密钥配置或关闭创建邮箱后重试"
	}

	return fmt.Sprintf("DuckMail 邮箱创建失败：%s", upstreamMessage)
}

func (s *AccountService) UpdateWarranty(ctx context.Context, parentID, warrantyID uint64, input AccountInput) (AccountRecord, error) {
	parent, err := s.findAccount(ctx, parentID)
	if err != nil {
		return AccountRecord{}, err
	}

	if err := validateWarrantyParent(parent); err != nil {
		return AccountRecord{}, err
	}

	return s.updateChildAccount(ctx, parentID, warrantyID, input, model.AccountRelationTypeWarranty, "warranty_not_found", "warranty account not found", "warranty_update_failed", "failed to update warranty account")
}

func (s *AccountService) UpdateSubAccount(ctx context.Context, parentID, subAccountID uint64, input AccountInput) (AccountRecord, error) {
	parent, err := s.findAccount(ctx, parentID)
	if err != nil {
		return AccountRecord{}, err
	}

	if err := validateSubAccountParent(parent); err != nil {
		return AccountRecord{}, err
	}

	return s.updateChildAccount(ctx, parentID, subAccountID, input, model.AccountRelationTypeSubAccount, "subaccount_not_found", "sub account not found", "subaccount_update_failed", "failed to update sub account")
}

func (s *AccountService) DeleteWarranty(ctx context.Context, parentID, warrantyID uint64) error {
	parent, err := s.findAccount(ctx, parentID)
	if err != nil {
		return err
	}

	if err := validateWarrantyParent(parent); err != nil {
		return err
	}

	return s.deleteChildAccount(ctx, parentID, warrantyID, model.AccountRelationTypeWarranty, "warranty_not_found", "warranty account not found", "warranty_delete_failed", "failed to delete warranty account")
}

func (s *AccountService) DeleteSubAccount(ctx context.Context, parentID, subAccountID uint64) error {
	parent, err := s.findAccount(ctx, parentID)
	if err != nil {
		return err
	}

	if err := validateSubAccountParent(parent); err != nil {
		return err
	}

	return s.deleteChildAccount(ctx, parentID, subAccountID, model.AccountRelationTypeSubAccount, "subaccount_not_found", "sub account not found", "subaccount_delete_failed", "failed to delete sub account")
}

func (s *AccountService) findAccount(ctx context.Context, id uint64) (model.Account, error) {
	account, err := s.accounts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Account{}, apperr.NotFound("account_not_found", "account not found")
		}

		return model.Account{}, apperr.Internal("account_lookup_failed", "failed to load account", err)
	}

	return account, nil
}

func validateAccountInput(input AccountInput) error {
	if strings.TrimSpace(input.Account) == "" {
		return apperr.BadRequest("account_required", "account is required")
	}

	if input.Password == "" {
		return apperr.BadRequest("password_required", "password is required")
	}

	switch input.Type {
	case model.AccountTypePlus, model.AccountTypeBusiness, model.AccountTypeCodex:
	default:
		return apperr.BadRequest("invalid_account_type", "type must be plus, business or codex")
	}

	switch input.Status {
	case model.AccountStatusNormal, model.AccountStatusBlocked:
	default:
		return apperr.BadRequest("invalid_account_status", "status must be normal or blocked")
	}

	return nil
}

func (s *AccountService) toRecords(accounts []model.Account) ([]AccountRecord, error) {
	items := make([]AccountRecord, 0, len(accounts))
	for _, account := range accounts {
		record, err := s.toRecord(account)
		if err != nil {
			return nil, err
		}
		items = append(items, record)
	}

	return items, nil
}

func (s *AccountService) createChildAccount(ctx context.Context, parentID uint64, input AccountInput, relationType model.AccountRelationType, errorCode, errorMessage string) (AccountRecord, error) {
	if err := validateAccountInput(input); err != nil {
		return AccountRecord{}, err
	}

	if input.CreateMailbox {
		if err := s.createDuckmailAccountIfNeeded(ctx, strings.TrimSpace(input.Account), input.Password); err != nil {
			return AccountRecord{}, err
		}
	}

	encrypted, err := s.cipher.Encrypt(input.Password)
	if err != nil {
		return AccountRecord{}, err
	}

	account := model.Account{
		Account:            strings.TrimSpace(input.Account),
		PasswordCiphertext: encrypted,
		Type:               input.Type,
		StartTime:          input.StartTime,
		EndTime:            input.EndTime,
		Status:             input.Status,
		Remark:             strings.TrimSpace(input.Remark),
		ParentID:           &parentID,
		RelationType:       relationType,
	}

	if err := s.accounts.Create(ctx, &account); err != nil {
		return AccountRecord{}, apperr.Internal(errorCode, errorMessage, err)
	}

	return s.toRecord(account)
}

func (s *AccountService) updateChildAccount(ctx context.Context, parentID, accountID uint64, input AccountInput, relationType model.AccountRelationType, notFoundCode, notFoundMessage, errorCode, errorMessage string) (AccountRecord, error) {
	if err := validateAccountInput(input); err != nil {
		return AccountRecord{}, err
	}

	account, err := s.findAccount(ctx, accountID)
	if err != nil {
		return AccountRecord{}, err
	}

	if account.ParentID == nil || *account.ParentID != parentID || !matchesRelationType(account.RelationType, relationType) {
		return AccountRecord{}, apperr.NotFound(notFoundCode, notFoundMessage)
	}

	encrypted, err := s.cipher.Encrypt(input.Password)
	if err != nil {
		return AccountRecord{}, err
	}

	account.Account = strings.TrimSpace(input.Account)
	account.PasswordCiphertext = encrypted
	account.Type = input.Type
	account.StartTime = input.StartTime
	account.EndTime = input.EndTime
	account.Status = input.Status
	account.Remark = strings.TrimSpace(input.Remark)

	if err := s.accounts.Save(ctx, &account); err != nil {
		return AccountRecord{}, apperr.Internal(errorCode, errorMessage, err)
	}

	return s.toRecord(account)
}

func (s *AccountService) deleteChildAccount(ctx context.Context, parentID, accountID uint64, relationType model.AccountRelationType, notFoundCode, notFoundMessage, errorCode, errorMessage string) error {
	account, err := s.findAccount(ctx, accountID)
	if err != nil {
		return err
	}

	if account.ParentID == nil || *account.ParentID != parentID || !matchesRelationType(account.RelationType, relationType) {
		return apperr.NotFound(notFoundCode, notFoundMessage)
	}

	if err := s.accounts.DeleteCascade(ctx, accountID); err != nil {
		return apperr.Internal(errorCode, errorMessage, err)
	}

	return nil
}

func validateWarrantyParent(parent model.Account) error {
	if parent.ParentID != nil {
		return apperr.BadRequest("account_not_parent", "warranty accounts can only belong to top-level accounts")
	}

	if parent.Status != model.AccountStatusBlocked {
		return apperr.Conflict("warranty_parent_not_blocked", "only blocked accounts can own warranty accounts")
	}

	return nil
}

func validateSubAccountParent(parent model.Account) error {
	if parent.ParentID != nil {
		return apperr.BadRequest("account_not_parent", "sub accounts can only belong to top-level accounts")
	}

	if parent.Type != model.AccountTypeCodex {
		return apperr.Conflict("subaccount_parent_not_codex", "only codex accounts can own sub accounts")
	}

	return nil
}

func matchesRelationType(actual, expected model.AccountRelationType) bool {
	if actual == expected {
		return true
	}

	return expected == model.AccountRelationTypeWarranty && actual == ""
}

func (s *AccountService) toRecord(account model.Account) (AccountRecord, error) {
	password, err := s.cipher.Decrypt(account.PasswordCiphertext)
	if err != nil {
		return AccountRecord{}, err
	}

	return AccountRecord{
		ID:             account.ID,
		Account:        account.Account,
		Password:       password,
		MaskedPassword: maskPassword(password),
		Type:           account.Type,
		StartTime:      account.StartTime,
		EndTime:        account.EndTime,
		Status:         account.Status,
		Remark:         account.Remark,
		ParentID:       account.ParentID,
		CreatedAt:      account.CreatedAt,
		UpdatedAt:      account.UpdatedAt,
	}, nil
}

func maskPassword(password string) string {
	if password == "" {
		return ""
	}

	if len(password) <= 2 {
		return strings.Repeat("*", len(password))
	}

	return password[:1] + strings.Repeat("*", len(password)-2) + password[len(password)-1:]
}

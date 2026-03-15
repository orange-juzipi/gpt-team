package service

import (
	"context"
	"testing"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/integration/mailbox"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"

	"gorm.io/gorm"
)

func TestBlockedAccountCannotReturnToNormalWithWarranties(t *testing.T) {
	t.Parallel()

	db, svc := newTestAccountService(t)
	_ = db

	parent := createAccount(t, svc, AccountInput{
		Account:  "owner@example.com",
		Password: "password123",
		Type:     model.AccountTypePlus,
		Status:   model.AccountStatusBlocked,
	})

	if _, err := svc.CreateWarranty(context.Background(), parent.ID, AccountInput{
		Account:  "warranty@example.com",
		Password: "password456",
		Type:     model.AccountTypePlus,
		Status:   model.AccountStatusNormal,
	}); err != nil {
		t.Fatalf("create warranty: %v", err)
	}

	if _, err := svc.Update(context.Background(), parent.ID, AccountInput{
		Account:  "owner@example.com",
		Password: "password123",
		Type:     model.AccountTypePlus,
		Status:   model.AccountStatusNormal,
	}); err == nil {
		t.Fatalf("expected conflict when moving blocked account with warranties back to normal")
	}
}

func TestWarrantyRequiresBlockedParent(t *testing.T) {
	t.Parallel()

	_, svc := newTestAccountService(t)
	parent := createAccount(t, svc, AccountInput{
		Account:  "owner@example.com",
		Password: "password123",
		Type:     model.AccountTypeBusiness,
		Status:   model.AccountStatusNormal,
	})

	if _, err := svc.CreateWarranty(context.Background(), parent.ID, AccountInput{
		Account:  "warranty@example.com",
		Password: "password456",
		Type:     model.AccountTypePlus,
		Status:   model.AccountStatusNormal,
	}); err == nil {
		t.Fatalf("expected blocked parent requirement")
	}
}

func TestDeleteAccountSoftDeletesChildren(t *testing.T) {
	t.Parallel()

	db, svc := newTestAccountService(t)
	parent := createAccount(t, svc, AccountInput{
		Account:  "owner@example.com",
		Password: "password123",
		Type:     model.AccountTypePlus,
		Status:   model.AccountStatusBlocked,
	})

	if _, err := svc.CreateWarranty(context.Background(), parent.ID, AccountInput{
		Account:  "warranty@example.com",
		Password: "password456",
		Type:     model.AccountTypePlus,
		Status:   model.AccountStatusNormal,
	}); err != nil {
		t.Fatalf("create warranty: %v", err)
	}

	if err := svc.Delete(context.Background(), parent.ID); err != nil {
		t.Fatalf("delete parent: %v", err)
	}

	var count int64
	if err := db.Unscoped().Model(&model.Account{}).Count(&count).Error; err != nil {
		t.Fatalf("count accounts: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 rows to remain in unscoped table, got %d", count)
	}

	var deletedCount int64
	if err := db.Unscoped().Model(&model.Account{}).Where("deleted_at IS NOT NULL").Count(&deletedCount).Error; err != nil {
		t.Fatalf("count deleted accounts: %v", err)
	}

	if deletedCount != 2 {
		t.Fatalf("expected both parent and child soft deleted, got %d", deletedCount)
	}
}

func TestListEmailsUsesAccountAddress(t *testing.T) {
	t.Parallel()

	_, svc := newTestAccountService(t)
	account := createAccount(t, svc, AccountInput{
		Account:  "mailbox@mail.example",
		Password: "password123",
		Type:     model.AccountTypeBusiness,
		Status:   model.AccountStatusNormal,
	})

	result, err := svc.ListEmails(context.Background(), account.ID)
	if err != nil {
		t.Fatalf("ListEmails: %v", err)
	}

	if result.Account != "mailbox@mail.example" {
		t.Fatalf("unexpected account: %s", result.Account)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 email, got %d", len(result.Items))
	}

	if result.Items[0].Subject != "Welcome" {
		t.Fatalf("unexpected subject: %s", result.Items[0].Subject)
	}
}

func TestListEmailsUsesConfiguredMailboxCredentialsWhenMatched(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	accountRepo := repository.NewAccountRepository(db)
	mailboxRepo := repository.NewMailboxProviderRepository(db)
	cipher, err := NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	mailboxSvc := NewMailboxProviderService(mailboxRepo, cipher)
	if _, err := mailboxSvc.Create(context.Background(), MailboxProviderInput{
		DomainSuffix: "mail.example",
		AccountEmail: "admin@mail.example",
		Password:     "admin-secret",
		Remark:       "test",
	}); err != nil {
		t.Fatalf("create mailbox provider: %v", err)
	}

	mailer := &captureAccountMailClient{}
	svc := NewAccountService(accountRepo, cipher, mailer, nil, mailboxSvc)
	account := createAccount(t, svc, AccountInput{
		Account:  "mailbox@mail.example",
		Password: "password123",
		Type:     model.AccountTypeBusiness,
		Status:   model.AccountStatusNormal,
	})

	if _, err := svc.ListEmails(context.Background(), account.ID); err != nil {
		t.Fatalf("ListEmails: %v", err)
	}

	if mailer.lastAuthEmail != "admin@mail.example" {
		t.Fatalf("expected configured auth email, got %q", mailer.lastAuthEmail)
	}
	if mailer.lastAuthPassword != "admin-secret" {
		t.Fatalf("expected configured auth password, got %q", mailer.lastAuthPassword)
	}
}

func TestListEmailsUsesDuckmailClientForDuckmailSuffix(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	accountRepo := repository.NewAccountRepository(db)
	mailboxRepo := repository.NewMailboxProviderRepository(db)
	cipher, err := NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	mailboxSvc := NewMailboxProviderService(mailboxRepo, cipher)
	if _, err := mailboxSvc.Create(context.Background(), MailboxProviderInput{
		ProviderType: model.MailboxProviderTypeDuckmail,
		DomainSuffix: "duckmail.sbs",
		Password:     "dk_example_secret",
		Remark:       "duck",
	}); err != nil {
		t.Fatalf("create mailbox provider: %v", err)
	}

	cloudMailer := &captureAccountMailClient{}
	duckMailer := &captureAccountMailClient{}
	svc := NewAccountService(accountRepo, cipher, cloudMailer, duckMailer, mailboxSvc)
	account := createAccount(t, svc, AccountInput{
		Account:  "mailbox@duckmail.sbs",
		Password: "duck-secret",
		Type:     model.AccountTypeBusiness,
		Status:   model.AccountStatusNormal,
	})

	if _, err := svc.ListEmails(context.Background(), account.ID); err != nil {
		t.Fatalf("ListEmails: %v", err)
	}

	if cloudMailer.calls != 0 {
		t.Fatalf("expected cloud mailer to stay unused, got %d calls", cloudMailer.calls)
	}
	if duckMailer.calls != 1 {
		t.Fatalf("expected duck mailer to be called once, got %d", duckMailer.calls)
	}
	if duckMailer.lastAuthEmail != "mailbox@duckmail.sbs" {
		t.Fatalf("expected duckmail auth email to use account email, got %q", duckMailer.lastAuthEmail)
	}
	if duckMailer.lastAuthPassword != "duck-secret" {
		t.Fatalf("expected duckmail auth password to use account password, got %q", duckMailer.lastAuthPassword)
	}
}

func TestCreateUsesDuckmailProvisioningWhenSuffixMatched(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	accountRepo := repository.NewAccountRepository(db)
	mailboxRepo := repository.NewMailboxProviderRepository(db)
	cipher, err := NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	mailboxSvc := NewMailboxProviderService(mailboxRepo, cipher)
	if _, err := mailboxSvc.Create(context.Background(), MailboxProviderInput{
		ProviderType: model.MailboxProviderTypeDuckmail,
		DomainSuffix: "duckmail.sbs",
		Password:     "dk_example_secret",
		Remark:       "duck",
	}); err != nil {
		t.Fatalf("create mailbox provider: %v", err)
	}

	duckMailer := &captureAccountMailClient{}
	svc := NewAccountService(accountRepo, cipher, fakeAccountMailClient{}, duckMailer, mailboxSvc)
	account, err := svc.Create(context.Background(), AccountInput{
		Account:       "fresh@duckmail.sbs",
		Password:      "duck-secret",
		Type:          model.AccountTypeBusiness,
		Status:        model.AccountStatusNormal,
		CreateMailbox: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if account.Account != "fresh@duckmail.sbs" {
		t.Fatalf("unexpected account: %s", account.Account)
	}
	if duckMailer.createCalls != 1 {
		t.Fatalf("expected duckmail create to be called once, got %d", duckMailer.createCalls)
	}
	if duckMailer.lastCreateAPIKey != "dk_example_secret" {
		t.Fatalf("unexpected api key: %q", duckMailer.lastCreateAPIKey)
	}
	if duckMailer.lastCreateAddress != "fresh@duckmail.sbs" {
		t.Fatalf("unexpected create address: %q", duckMailer.lastCreateAddress)
	}
}

func TestCreateSkipsDuckmailProvisioningWhenDisabled(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	accountRepo := repository.NewAccountRepository(db)
	mailboxRepo := repository.NewMailboxProviderRepository(db)
	cipher, err := NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	mailboxSvc := NewMailboxProviderService(mailboxRepo, cipher)
	if _, err := mailboxSvc.Create(context.Background(), MailboxProviderInput{
		ProviderType: model.MailboxProviderTypeDuckmail,
		DomainSuffix: "duckmail.sbs",
		Password:     "dk_example_secret",
		Remark:       "duck",
	}); err != nil {
		t.Fatalf("create mailbox provider: %v", err)
	}

	duckMailer := &captureAccountMailClient{}
	svc := NewAccountService(accountRepo, cipher, fakeAccountMailClient{}, duckMailer, mailboxSvc)
	account, err := svc.Create(context.Background(), AccountInput{
		Account:       "skip@duckmail.sbs",
		Password:      "duck-secret",
		Type:          model.AccountTypeBusiness,
		Status:        model.AccountStatusNormal,
		CreateMailbox: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if account.Account != "skip@duckmail.sbs" {
		t.Fatalf("unexpected account: %s", account.Account)
	}
	if duckMailer.createCalls != 0 {
		t.Fatalf("expected duckmail create to be skipped, got %d calls", duckMailer.createCalls)
	}
}

func TestCreateWrapsDuckmailProvisioningErrors(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	accountRepo := repository.NewAccountRepository(db)
	mailboxRepo := repository.NewMailboxProviderRepository(db)
	cipher, err := NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	mailboxSvc := NewMailboxProviderService(mailboxRepo, cipher)
	if _, err := mailboxSvc.Create(context.Background(), MailboxProviderInput{
		ProviderType: model.MailboxProviderTypeDuckmail,
		DomainSuffix: "duckmail.sbs",
		Password:     "dk_example_secret",
		Remark:       "duck",
	}); err != nil {
		t.Fatalf("create mailbox provider: %v", err)
	}

	duckMailer := &captureAccountMailClient{
		createErr: apperr.Forbidden("duckmail_http_403", "invalid duckmail api key"),
	}
	svc := NewAccountService(accountRepo, cipher, fakeAccountMailClient{}, duckMailer, mailboxSvc)
	_, err = svc.Create(context.Background(), AccountInput{
		Account:       "error@duckmail.sbs",
		Password:      "duck-secret",
		Type:          model.AccountTypeBusiness,
		Status:        model.AccountStatusNormal,
		CreateMailbox: true,
	})
	if err == nil {
		t.Fatalf("expected create to fail")
	}

	if apperr.Status(err) != 502 {
		t.Fatalf("expected 502, got %d", apperr.Status(err))
	}
	if apperr.Code(err) != "duckmail_account_create_failed" {
		t.Fatalf("unexpected code: %s", apperr.Code(err))
	}
	if apperr.Message(err) != "DuckMail 邮箱创建失败，请检查邮箱管理中的密钥配置或关闭创建邮箱后重试" {
		t.Fatalf("unexpected message: %s", apperr.Message(err))
	}
}

func newTestAccountService(t *testing.T) (*gorm.DB, *AccountService) {
	t.Helper()

	db := newTestDB(t)
	repo := repository.NewAccountRepository(db)
	cipher, err := NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	return db, NewAccountService(repo, cipher, fakeAccountMailClient{}, nil, nil)
}

func createAccount(t *testing.T, svc *AccountService, input AccountInput) AccountRecord {
	t.Helper()

	account, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	return account
}

type fakeAccountMailClient struct{}

func (fakeAccountMailClient) ListInboxEmails(ctx context.Context, authEmail, authPassword, targetEmail string) ([]mailbox.Email, error) {
	return []mailbox.Email{
		{
			ID:         "1",
			Account:    targetEmail,
			From:       "noreply@example.com",
			FromName:   "Example",
			Subject:    "Welcome",
			Preview:    "Hello from the mailbox",
			ReceivedAt: "2026-03-15 10:20:30",
		},
	}, nil
}

type captureAccountMailClient struct {
	lastAuthEmail      string
	lastAuthPassword   string
	calls              int
	lastCreateAPIKey   string
	lastCreateAddress  string
	lastCreatePassword string
	createCalls        int
	createErr          error
}

func (c *captureAccountMailClient) ListInboxEmails(ctx context.Context, authEmail, authPassword, targetEmail string) ([]mailbox.Email, error) {
	c.lastAuthEmail = authEmail
	c.lastAuthPassword = authPassword
	c.calls++
	return []mailbox.Email{
		{
			ID:         "1",
			Account:    targetEmail,
			From:       "noreply@example.com",
			FromName:   "Example",
			Subject:    "Welcome",
			Preview:    "Hello from the mailbox",
			ReceivedAt: "2026-03-15 10:20:30",
		},
	}, nil
}

func (c *captureAccountMailClient) CreateAccount(ctx context.Context, apiKey, address, password string) error {
	c.lastCreateAPIKey = apiKey
	c.lastCreateAddress = address
	c.lastCreatePassword = password
	c.createCalls++
	return c.createErr
}

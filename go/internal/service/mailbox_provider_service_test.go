package service

import (
	"context"
	"testing"

	"gpt-team-api/internal/repository"
)

func TestMailboxProviderResolvesConfigByAccountDomain(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := repository.NewMailboxProviderRepository(db)
	cipher, err := NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	svc := NewMailboxProviderService(repo, cipher)
	if _, err := svc.Create(context.Background(), MailboxProviderInput{
		DomainSuffix: "@mail.example",
		AccountEmail: "admin@mail.example",
		Password:     "provider-password",
		Remark:       "primary",
	}); err != nil {
		t.Fatalf("create provider: %v", err)
	}

	config, matched, err := svc.ResolveConfigByAccount(context.Background(), "worker@mail.example")
	if err != nil {
		t.Fatalf("ResolveConfigByAccount: %v", err)
	}

	if !matched {
		t.Fatalf("expected matched provider")
	}

	if config.AccountEmail != "admin@mail.example" {
		t.Fatalf("unexpected account email: %s", config.AccountEmail)
	}
	if config.Password != "provider-password" {
		t.Fatalf("unexpected password: %s", config.Password)
	}
	if config.ProviderType != "cloudmail" {
		t.Fatalf("unexpected provider type: %s", config.ProviderType)
	}
}

func TestMailboxProviderAllowsDuckmailWithoutAdminCredentials(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := repository.NewMailboxProviderRepository(db)
	cipher, err := NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	svc := NewMailboxProviderService(repo, cipher)
	record, err := svc.Create(context.Background(), MailboxProviderInput{
		ProviderType: "duckmail",
		DomainSuffix: "duckmail.sbs",
		Password:     "dk_example_secret",
		Remark:       "duck",
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	if record.ProviderType != "duckmail" {
		t.Fatalf("unexpected provider type: %s", record.ProviderType)
	}
	if record.AccountEmail != "" {
		t.Fatalf("expected empty account email, got %q", record.AccountEmail)
	}
	if record.Password != "dk_example_secret" {
		t.Fatalf("unexpected stored api key: %q", record.Password)
	}
}

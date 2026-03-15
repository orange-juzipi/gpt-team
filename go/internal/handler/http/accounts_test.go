package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gpt-team-api/internal/integration/mailbox"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"
	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetAccountEmailsEndpoint(t *testing.T) {
	t.Parallel()

	router := newTestAccountRouter(t)
	request := httptest.NewRequest(http.MethodGet, "/api/accounts/1/emails", http.NoBody)
	attachAuthCookie(t, router, request)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	payload := decodeBody[struct {
		Data service.AccountEmailList `json:"data"`
	}](t, response.Body)

	if payload.Data.Account != "agent@mail.example" {
		t.Fatalf("unexpected account: %s", payload.Data.Account)
	}

	if len(payload.Data.Items) != 1 {
		t.Fatalf("expected 1 email, got %d", len(payload.Data.Items))
	}
}

func TestAccountWriteEndpointsRequireAuthenticatedUser(t *testing.T) {
	t.Parallel()

	router := newTestAccountRouter(t)

	createUserBody := bytes.NewBufferString(`{"username":"viewer","password":"viewer123","role":"member"}`)
	createUserRequest := httptest.NewRequest(http.MethodPost, "/api/users", createUserBody)
	createUserRequest.Header.Set("Content-Type", "application/json")
	attachAuthCookie(t, router, createUserRequest)
	createUserResponse := httptest.NewRecorder()
	router.ServeHTTP(createUserResponse, createUserRequest)

	if createUserResponse.Code != http.StatusCreated {
		t.Fatalf("expected user creation to succeed, got %d", createUserResponse.Code)
	}

	createAccountBody := bytes.NewBufferString(`{"account":"readonly@mail.example","password":"password123","type":"business","status":"normal","remark":""}`)
	createAccountRequest := httptest.NewRequest(http.MethodPost, "/api/accounts", createAccountBody)
	createAccountRequest.Header.Set("Content-Type", "application/json")
	attachAuthCookieForCredentials(t, router, createAccountRequest, "viewer", "viewer123")
	createAccountResponse := httptest.NewRecorder()
	router.ServeHTTP(createAccountResponse, createAccountRequest)

	if createAccountResponse.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createAccountResponse.Code)
	}
}

func newTestAccountRouter(t *testing.T) *gin.Engine {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/account-handler.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.Card{}, &model.CardEvent{}, &model.Account{}, &model.MailboxProvider{}, &model.User{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	accountRepo := repository.NewAccountRepository(db)
	mailboxProviderRepo := repository.NewMailboxProviderRepository(db)
	userRepo := repository.NewUserRepository(db)
	cipher, err := service.NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	mailboxProviderService := service.NewMailboxProviderService(mailboxProviderRepo, cipher)
	accountService := service.NewAccountService(accountRepo, cipher, fakeAccountHandlerMailClient{}, nil, mailboxProviderService)
	account, err := accountService.Create(context.Background(), service.AccountInput{
		Account:  "agent@mail.example",
		Password: "password123",
		Type:     model.AccountTypeBusiness,
		Status:   model.AccountStatusNormal,
	})
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if account.ID != 1 {
		t.Fatalf("expected seeded account id 1, got %d", account.ID)
	}

	userService := service.NewUserService(userRepo)
	if err := userService.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatalf("ensure default admin: %v", err)
	}

	return NewRouter(
		service.NewCardService(
			repository.NewCardRepository(db),
			repository.NewCardEventRepository(db),
			fakeHandlerEfuncard{},
			fakeHandlerProfile{},
		),
		accountService,
		service.NewProfileService(fakeHandlerProfile{}),
		mailboxProviderService,
		userService,
		service.NewAuthService(
			userRepo,
			service.NewSessionManager("12345678901234567890123456789012", 24*time.Hour),
		),
	)
}

type fakeAccountHandlerMailClient struct{}

func (fakeAccountHandlerMailClient) ListInboxEmails(ctx context.Context, authEmail, authPassword, targetEmail string) ([]mailbox.Email, error) {
	return []mailbox.Email{
		{
			ID:         "99",
			Account:    targetEmail,
			From:       "noreply@example.com",
			FromName:   "Example",
			Subject:    "Inbox ready",
			Preview:    "This mailbox has one no-owner email",
			ReceivedAt: "2026-03-15 10:20:30",
		},
	}, nil
}

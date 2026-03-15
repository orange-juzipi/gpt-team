package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/integration/efuncard"
	"gpt-team-api/internal/integration/mailbox"
	"gpt-team-api/internal/integration/meiguodizhi"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"
	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestImportCardsEndpoint(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, nil)
	body := bytes.NewBufferString(`{"rawText":"A\nA\nB","cardType":"es","cardLimit":2}`)
	request := httptest.NewRequest(http.MethodPost, "/api/cards/import", body)
	request.Header.Set("Content-Type", "application/json")
	attachAuthCookie(t, router, request)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}

func TestImportCardsEndpointRequiresCardLimit(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, nil)
	body := bytes.NewBufferString(`{"rawText":"A\nB","cardType":"us"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/cards/import", body)
	request.Header.Set("Content-Type", "application/json")
	attachAuthCookie(t, router, request)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestActivateCardConflictMapsTo409(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, func() service.EfuncardClient {
		return fakeHandlerEfuncard{
			redeemFn: func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error) {
				return efuncard.APIResponse[efuncard.RedeemData]{}, apperr.Conflict("efuncard_http_409", "already used")
			},
		}
	})

	request := httptest.NewRequest(http.MethodPost, "/api/cards/1/activate", http.NoBody)
	attachAuthCookie(t, router, request)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", response.Code)
	}
}

func TestCardsEndpointRequiresAdminRole(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, nil)

	createUserBody := bytes.NewBufferString(`{"username":"operator","password":"operator123","role":"member"}`)
	createUserRequest := httptest.NewRequest(http.MethodPost, "/api/users", createUserBody)
	createUserRequest.Header.Set("Content-Type", "application/json")
	attachAuthCookie(t, router, createUserRequest)
	createUserResponse := httptest.NewRecorder()
	router.ServeHTTP(createUserResponse, createUserRequest)

	if createUserResponse.Code != http.StatusCreated {
		t.Fatalf("expected user creation to succeed, got %d", createUserResponse.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/cards", http.NoBody)
	attachAuthCookieForCredentials(t, router, request, "operator", "operator123")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func TestRandomProfileEndpoint(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/profiles/random", http.NoBody)
	attachAuthCookie(t, router, request)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	payload := decodeBody[struct {
		Data service.RandomProfile `json:"data"`
	}](t, response.Body)

	if payload.Data.FullName != "Ada Lovelace" {
		t.Fatalf("unexpected full name: %s", payload.Data.FullName)
	}
	if payload.Data.Birthday != "1815-12-10" {
		t.Fatalf("unexpected birthday: %s", payload.Data.Birthday)
	}
}

type fakeHandlerEfuncard struct {
	redeemFn func(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error)
}

func (f fakeHandlerEfuncard) Redeem(ctx context.Context, code string) (efuncard.APIResponse[efuncard.RedeemData], error) {
	if f.redeemFn != nil {
		return f.redeemFn(ctx, code)
	}
	return efuncard.APIResponse[efuncard.RedeemData]{}, nil
}

func (f fakeHandlerEfuncard) QueryCard(ctx context.Context, code string) (efuncard.APIResponse[efuncard.QueryData], error) {
	return efuncard.APIResponse[efuncard.QueryData]{}, nil
}

func (f fakeHandlerEfuncard) Billing(ctx context.Context, code string) (efuncard.APIResponse[efuncard.BillingData], error) {
	return efuncard.APIResponse[efuncard.BillingData]{}, nil
}

func (f fakeHandlerEfuncard) ThreeDS(ctx context.Context, code string, minutes int) (efuncard.APIResponse[efuncard.ThreeDSData], error) {
	return efuncard.APIResponse[efuncard.ThreeDSData]{}, nil
}

type fakeHandlerProfile struct{}

func (fakeHandlerProfile) FetchProfile(ctx context.Context) (meiguodizhi.ProfileResponse, error) {
	return meiguodizhi.ProfileResponse{
		FullName: "Ada Lovelace",
		Birthday: "1815-12-10",
		Raw:      `{"fullName":"Ada Lovelace","birthday":"1815-12-10"}`,
	}, nil
}

func newTestRouter(t *testing.T, clientFactory func() service.EfuncardClient) *gin.Engine {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/handler.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.Card{}, &model.CardEvent{}, &model.Account{}, &model.MailboxProvider{}, &model.User{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	cardRepo := repository.NewCardRepository(db)
	if err := cardRepo.CreateMany(context.Background(), []model.Card{
		{Code: "CDK-1", CardType: model.CardTypeUS, CardLimit: 1, Status: model.CardStatusUnactivated},
	}); err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	eventRepo := repository.NewCardEventRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	mailboxProviderRepo := repository.NewMailboxProviderRepository(db)
	userRepo := repository.NewUserRepository(db)
	cipher, err := service.NewCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	client := service.EfuncardClient(fakeHandlerEfuncard{})
	if clientFactory != nil {
		client = clientFactory()
	}

	cardService := service.NewCardService(cardRepo, eventRepo, client, fakeHandlerProfile{})
	mailboxProviderService := service.NewMailboxProviderService(mailboxProviderRepo, cipher)
	accountService := service.NewAccountService(accountRepo, cipher, fakeHandlerMailClient{}, nil, mailboxProviderService)
	userService := service.NewUserService(userRepo)
	if err := userService.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatalf("ensure default admin: %v", err)
	}
	authService := service.NewAuthService(
		userRepo,
		service.NewSessionManager("12345678901234567890123456789012", 24*time.Hour),
	)
	return NewRouter(
		cardService,
		accountService,
		service.NewProfileService(fakeHandlerProfile{}),
		mailboxProviderService,
		userService,
		authService,
	)
}

func decodeBody[T any](t *testing.T, body *bytes.Buffer) T {
	t.Helper()
	var payload T
	if err := json.Unmarshal(body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return payload
}

type fakeHandlerMailClient struct{}

func (fakeHandlerMailClient) ListInboxEmails(ctx context.Context, authEmail, authPassword, targetEmail string) ([]mailbox.Email, error) {
	return []mailbox.Email{
		{
			ID:         "1",
			Account:    targetEmail,
			From:       "noreply@example.com",
			FromName:   "Example",
			Subject:    "Hello",
			Preview:    "Preview",
			ReceivedAt: "2026-03-15 10:20:30",
		},
	}, nil
}

func attachAuthCookie(t *testing.T, router *gin.Engine, request *http.Request) {
	t.Helper()

	attachAuthCookieForCredentials(t, router, request, "admin", "admin")
}

func attachAuthCookieForCredentials(t *testing.T, router *gin.Engine, request *http.Request, username, password string) {
	t.Helper()

	loginBody := bytes.NewBufferString(
		`{"username":"` + username + `","password":"` + password + `"}`,
	)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login failed with status %d", loginResponse.Code)
	}

	for _, cookie := range loginResponse.Result().Cookies() {
		request.AddCookie(cookie)
	}
}

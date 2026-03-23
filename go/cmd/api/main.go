package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"gpt-team-api/internal/config"
	"gpt-team-api/internal/db"
	httpapi "gpt-team-api/internal/handler/http"
	"gpt-team-api/internal/integration/cloudmail"
	"gpt-team-api/internal/integration/duckmail"
	"gpt-team-api/internal/integration/efuncard"
	"gpt-team-api/internal/integration/meiguodizhi"
	"gpt-team-api/internal/repository"
	"gpt-team-api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{Timeout: cfg.HTTPTimeout}
	duckmailHTTPClient := &http.Client{Timeout: cfg.DuckmailHTTPTimeout}

	cardRepo := repository.NewCardRepository(database)
	eventRepo := repository.NewCardEventRepository(database)
	accountRepo := repository.NewAccountRepository(database)
	mailboxProviderRepo := repository.NewMailboxProviderRepository(database)
	userRepo := repository.NewUserRepository(database)

	cipher, err := service.NewCipher(cfg.AccountEncryptionKey)
	if err != nil {
		log.Fatal(err)
	}

	mailboxProviderService := service.NewMailboxProviderService(mailboxProviderRepo, cipher)
	cardService := service.NewCardService(
		cardRepo,
		eventRepo,
		efuncard.NewClient(cfg.EfuncardBaseURL, cfg.EfuncardAPIKey, httpClient),
		meiguodizhi.NewClient(cfg.MeiguodizhiBaseURL, httpClient),
	)
	profileService := service.NewProfileService(
		meiguodizhi.NewClient(cfg.MeiguodizhiBaseURL, httpClient),
	)
	accountService := service.NewAccountService(
		accountRepo,
		cipher,
		cloudmail.NewClient(cfg.CloudmailBaseURL, cfg.CloudmailAPIToken, httpClient),
		duckmail.NewClient(cfg.DuckmailBaseURL, duckmailHTTPClient),
		mailboxProviderService,
	)
	userService := service.NewUserService(userRepo)
	if err := userService.EnsureDefaultAdmin(context.Background()); err != nil {
		log.Fatal(err)
	}

	authService := service.NewAuthService(
		userRepo,
		service.NewSessionManager(cfg.AccountEncryptionKey, 7*24*time.Hour),
	)

	router := httpapi.NewRouter(
		cardService,
		accountService,
		profileService,
		mailboxProviderService,
		userService,
		authService,
	)
	log.Fatal(router.Run(":" + cfg.Port))
}

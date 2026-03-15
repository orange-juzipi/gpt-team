package httpapi

import (
	"net/http"

	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	cardService *service.CardService,
	accountService *service.AccountService,
	profileService *service.ProfileService,
	mailboxProviderService *service.MailboxProviderService,
	userService *service.UserService,
	authService *service.AuthService,
) *gin.Engine {
	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}
	router.Use(gin.Logger(), gin.Recovery())

	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		respond(c, http.StatusOK, gin.H{"status": "ok"})
	})

	authMiddleware := NewAuthMiddleware(authService)

	NewAuthHandler(authService, authMiddleware).Register(api)

	protected := api.Group("")
	protected.Use(authMiddleware.RequireAuth())
	accountHandler := NewAccountHandler(accountService)
	profileHandler := NewProfileHandler(profileService)
	protected.GET("/accounts", accountHandler.listAccounts)
	protected.GET("/accounts/:id/emails", accountHandler.listEmails)
	protected.GET("/accounts/:id/warranties", accountHandler.listWarranties)
	protected.POST("/accounts", accountHandler.createAccount)
	protected.PUT("/accounts/:id", accountHandler.updateAccount)
	protected.DELETE("/accounts/:id", accountHandler.deleteAccount)
	protected.POST("/accounts/:id/warranties", accountHandler.createWarranty)
	protected.PUT("/accounts/:id/warranties/:wid", accountHandler.updateWarranty)
	protected.DELETE("/accounts/:id/warranties/:wid", accountHandler.deleteWarranty)
	profileHandler.Register(protected)

	adminOnly := protected.Group("")
	adminOnly.Use(authMiddleware.RequireAdmin())
	NewCardHandler(cardService).Register(adminOnly)
	NewMailboxProviderHandler(mailboxProviderService).Register(adminOnly)
	NewUserHandler(userService).Register(adminOnly)

	return router
}

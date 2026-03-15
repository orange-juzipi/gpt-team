package httpapi

import (
	"net/http"
	"strings"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "gpt_team_session"

const authUserContextKey = "auth_user"

type AuthHandler struct {
	service    *service.AuthService
	middleware *AuthMiddleware
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthHandler(service *service.AuthService, middleware *AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		service:    service,
		middleware: middleware,
	}
}

func (h *AuthHandler) Register(group *gin.RouterGroup) {
	auth := group.Group("/auth")
	auth.POST("/login", h.login)

	protected := auth.Group("")
	protected.Use(h.middleware.RequireAuth())
	protected.GET("/me", h.me)
	protected.POST("/logout", h.logout)
}

func (h *AuthHandler) login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperr.BadRequest("invalid_login_payload", "username and password are required"))
		return
	}

	result, err := h.service.Login(c.Request.Context(), service.LoginInput{
		Username: strings.TrimSpace(request.Username),
		Password: request.Password,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	writeSessionCookie(c, result.Token)
	respond(c, http.StatusOK, result.User)
}

func (h *AuthHandler) me(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		respondError(c, apperr.Unauthorized("invalid_session", "login required"))
		return
	}

	respond(c, http.StatusOK, user)
}

func (h *AuthHandler) logout(c *gin.Context) {
	clearSessionCookie(c)
	respondNoContent(c)
}

func writeSessionCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   7 * 24 * 60 * 60,
		Secure:   requestIsSecure(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Secure:   requestIsSecure(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func requestIsSecure(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}

	return strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
}

func currentUser(c *gin.Context) (service.UserRecord, bool) {
	value, ok := c.Get(authUserContextKey)
	if !ok {
		return service.UserRecord{}, false
	}

	user, ok := value.(service.UserRecord)
	return user, ok
}

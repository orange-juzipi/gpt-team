package httpapi

import (
	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	service *service.AuthService
}

func NewAuthMiddleware(service *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{service: service}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessionCookieName)
		if err != nil || token == "" {
			respondError(c, apperr.Unauthorized("invalid_session", "login required"))
			c.Abort()
			return
		}

		user, err := m.service.CurrentUser(c.Request.Context(), token)
		if err != nil {
			respondError(c, err)
			c.Abort()
			return
		}

		c.Set(authUserContextKey, user)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			respondError(c, apperr.Unauthorized("invalid_session", "login required"))
			c.Abort()
			return
		}

		if user.Role != model.UserRoleAdmin {
			respondError(c, apperr.Forbidden("admin_required", "admin access required"))
			c.Abort()
			return
		}

		c.Next()
	}
}

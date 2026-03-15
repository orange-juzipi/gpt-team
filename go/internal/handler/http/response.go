package httpapi

import (
	"gpt-team-api/internal/apperr"

	"github.com/gin-gonic/gin"
)

func respond(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"data": data,
	})
}

func respondError(c *gin.Context, err error) {
	c.JSON(apperr.Status(err), gin.H{
		"error": gin.H{
			"code":    apperr.Code(err),
			"message": apperr.Message(err),
		},
	})
}

func respondNoContent(c *gin.Context) {
	c.Status(204)
}

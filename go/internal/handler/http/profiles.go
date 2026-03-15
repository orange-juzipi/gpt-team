package httpapi

import (
	"net/http"

	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	service *service.ProfileService
}

func NewProfileHandler(service *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

func (h *ProfileHandler) Register(group *gin.RouterGroup) {
	group.GET("/profiles/random", h.fetchRandomProfile)
}

func (h *ProfileHandler) fetchRandomProfile(c *gin.Context) {
	result, err := h.service.FetchRandomProfile(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

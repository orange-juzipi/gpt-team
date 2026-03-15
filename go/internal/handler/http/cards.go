package httpapi

import (
	"net/http"
	"strconv"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	service *service.CardService
}

type importCardsRequest struct {
	RawText   string `json:"rawText"`
	CardType  string `json:"cardType"`
	CardLimit *int   `json:"cardLimit"`
}

type threeDSRequest struct {
	Minutes int `json:"minutes"`
}

func NewCardHandler(service *service.CardService) *CardHandler {
	return &CardHandler{service: service}
}

func (h *CardHandler) Register(group *gin.RouterGroup) {
	group.POST("/cards/import", h.importCards)
	group.GET("/cards", h.listCards)
	group.GET("/cards/:id", h.getCard)
	group.POST("/cards/:id/activate", h.activateCard)
	group.POST("/cards/:id/query", h.queryCard)
	group.POST("/cards/:id/billing", h.getBilling)
	group.POST("/cards/:id/3ds", h.getThreeDS)
	group.POST("/cards/:id/profile/refresh", h.refreshProfile)
	group.DELETE("/cards/:id", h.deleteCard)
}

func (h *CardHandler) importCards(c *gin.Context) {
	var request importCardsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperr.BadRequest("invalid_import_payload", "rawText, cardType, and cardLimit are required"))
		return
	}
	if request.CardLimit == nil {
		respondError(c, apperr.BadRequest("invalid_import_payload", "rawText, cardType, and cardLimit are required"))
		return
	}

	result, err := h.service.Import(c.Request.Context(), service.ImportCardsInput{
		RawText:   request.RawText,
		CardType:  model.CardType(request.CardType),
		CardLimit: *request.CardLimit,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusCreated, result)
}

func (h *CardHandler) listCards(c *gin.Context) {
	result, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *CardHandler) getCard(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.Detail(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *CardHandler) activateCard(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.Activate(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *CardHandler) queryCard(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.Query(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *CardHandler) getBilling(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.Billing(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *CardHandler) getThreeDS(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	var request threeDSRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperr.BadRequest("invalid_three_ds_payload", "minutes is required"))
		return
	}

	result, err := h.service.ThreeDS(c.Request.Context(), id, request.Minutes)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *CardHandler) refreshProfile(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.RefreshProfile(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *CardHandler) deleteCard(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}

	respondNoContent(c)
}

func parseID(raw string) (uint64, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, apperr.BadRequest("invalid_id", "id must be a positive integer")
	}

	return id, nil
}

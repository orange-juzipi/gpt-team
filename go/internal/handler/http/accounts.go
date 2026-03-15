package httpapi

import (
	"net/http"
	"time"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	service *service.AccountService
}

type accountRequest struct {
	Account       string  `json:"account"`
	Password      string  `json:"password"`
	Type          string  `json:"type"`
	StartTime     *string `json:"startTime"`
	EndTime       *string `json:"endTime"`
	Status        string  `json:"status"`
	Remark        string  `json:"remark"`
	CreateMailbox *bool   `json:"createMailbox"`
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) Register(group *gin.RouterGroup) {
	group.GET("/accounts", h.listAccounts)
	group.GET("/accounts/:id/emails", h.listEmails)
	group.POST("/accounts", h.createAccount)
	group.PUT("/accounts/:id", h.updateAccount)
	group.DELETE("/accounts/:id", h.deleteAccount)
	group.GET("/accounts/:id/warranties", h.listWarranties)
	group.POST("/accounts/:id/warranties", h.createWarranty)
	group.PUT("/accounts/:id/warranties/:wid", h.updateWarranty)
	group.DELETE("/accounts/:id/warranties/:wid", h.deleteWarranty)
}

func (h *AccountHandler) listAccounts(c *gin.Context) {
	result, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *AccountHandler) listEmails(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.ListEmails(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *AccountHandler) createAccount(c *gin.Context) {
	input, err := bindAccountInput(c)
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusCreated, result)
}

func (h *AccountHandler) updateAccount(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	input, err := bindAccountInput(c)
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *AccountHandler) deleteAccount(c *gin.Context) {
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

func (h *AccountHandler) listWarranties(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.ListWarranties(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *AccountHandler) createWarranty(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	input, err := bindAccountInput(c)
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.CreateWarranty(c.Request.Context(), id, input)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusCreated, result)
}

func (h *AccountHandler) updateWarranty(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	warrantyID, err := parseID(c.Param("wid"))
	if err != nil {
		respondError(c, err)
		return
	}

	input, err := bindAccountInput(c)
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.service.UpdateWarranty(c.Request.Context(), id, warrantyID, input)
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *AccountHandler) deleteWarranty(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	warrantyID, err := parseID(c.Param("wid"))
	if err != nil {
		respondError(c, err)
		return
	}

	if err := h.service.DeleteWarranty(c.Request.Context(), id, warrantyID); err != nil {
		respondError(c, err)
		return
	}

	respondNoContent(c)
}

func bindAccountInput(c *gin.Context) (service.AccountInput, error) {
	var request accountRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		return service.AccountInput{}, apperr.BadRequest("invalid_account_payload", "account payload is invalid")
	}

	startTime, err := parseOptionalTime(request.StartTime)
	if err != nil {
		return service.AccountInput{}, err
	}

	endTime, err := parseOptionalTime(request.EndTime)
	if err != nil {
		return service.AccountInput{}, err
	}

	return service.AccountInput{
		Account:       request.Account,
		Password:      request.Password,
		Type:          model.AccountType(request.Type),
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        model.AccountStatus(request.Status),
		Remark:        request.Remark,
		CreateMailbox: resolveCreateMailbox(request.CreateMailbox),
	}, nil
}

func resolveCreateMailbox(value *bool) bool {
	if value == nil {
		return true
	}

	return *value
}

func parseOptionalTime(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return nil, apperr.BadRequest("invalid_time", "time must use RFC3339 format")
	}

	return &parsed, nil
}

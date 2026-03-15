package httpapi

import (
	"net/http"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
)

type MailboxProviderHandler struct {
	service *service.MailboxProviderService
}

type mailboxProviderRequest struct {
	ProviderType string `json:"providerType"`
	DomainSuffix string `json:"domainSuffix"`
	AccountEmail string `json:"accountEmail"`
	Password     string `json:"password"`
	Remark       string `json:"remark"`
}

func NewMailboxProviderHandler(service *service.MailboxProviderService) *MailboxProviderHandler {
	return &MailboxProviderHandler{service: service}
}

func (h *MailboxProviderHandler) Register(group *gin.RouterGroup) {
	group.GET("/mailbox-providers", h.listProviders)
	group.POST("/mailbox-providers", h.createProvider)
	group.PUT("/mailbox-providers/:id", h.updateProvider)
	group.DELETE("/mailbox-providers/:id", h.deleteProvider)
}

func (h *MailboxProviderHandler) listProviders(c *gin.Context) {
	result, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *MailboxProviderHandler) createProvider(c *gin.Context) {
	input, err := bindMailboxProviderInput(c)
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

func (h *MailboxProviderHandler) updateProvider(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	input, err := bindMailboxProviderInput(c)
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

func (h *MailboxProviderHandler) deleteProvider(c *gin.Context) {
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

func bindMailboxProviderInput(c *gin.Context) (service.MailboxProviderInput, error) {
	var request mailboxProviderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		return service.MailboxProviderInput{}, apperr.BadRequest("invalid_mailbox_provider_payload", "providerType and domainSuffix are required")
	}

	return service.MailboxProviderInput{
		ProviderType: model.MailboxProviderType(request.ProviderType),
		DomainSuffix: request.DomainSuffix,
		AccountEmail: request.AccountEmail,
		Password:     request.Password,
		Remark:       request.Remark,
	}, nil
}

package httpapi

import (
	"net/http"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *service.UserService
}

type userRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(group *gin.RouterGroup) {
	group.GET("/users", h.listUsers)
	group.POST("/users", h.createUser)
	group.PUT("/users/:id", h.updateUser)
	group.DELETE("/users/:id", h.deleteUser)
}

func (h *UserHandler) listUsers(c *gin.Context) {
	result, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, result)
}

func (h *UserHandler) createUser(c *gin.Context) {
	input, err := bindUserInput(c)
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

func (h *UserHandler) updateUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	input, err := bindUserInput(c)
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

func (h *UserHandler) deleteUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	user, ok := currentUser(c)
	if !ok {
		respondError(c, apperr.Unauthorized("invalid_session", "login required"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id, user.ID); err != nil {
		respondError(c, err)
		return
	}

	respondNoContent(c)
}

func bindUserInput(c *gin.Context) (service.UserInput, error) {
	var request userRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		return service.UserInput{}, apperr.BadRequest("invalid_user_payload", "username, password, and role are required")
	}

	return service.UserInput{
		Username: request.Username,
		Password: request.Password,
		Role:     model.UserRole(request.Role),
	}, nil
}

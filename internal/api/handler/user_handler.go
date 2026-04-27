package handler

import (
	"net/http"
	"strings"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUC *usecase.UserUseCase
}

func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update user id is required")
		return
	}

	var req request.ResquestUserDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update user request body invalid")
		return
	}

	if strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.PasswordHash) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update user has no fields to update")
		return
	}

	updatedUser, err := h.userUC.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		logAndRespondWithError(c, err, "update user failed")
		return
	}

	Success(c, http.StatusOK, toUserDTO(updatedUser))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "delete user id is required")
		return
	}

	if err := h.userUC.DeleteUser(c.Request.Context(), id); err != nil {
		logAndRespondWithError(c, err, "delete user failed")
		return
	}

	c.Status(http.StatusNoContent)
}

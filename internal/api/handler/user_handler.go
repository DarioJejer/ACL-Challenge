package handler

import (
	"net/http"
	"strings"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/usecase"

	"github.com/gin-gonic/gin"
)

func UpdateUser(c *gin.Context) {
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

	Success(c, http.StatusOK, gin.H{"message": "stub: updated"})
}

func DeleteUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "delete user id is required")
		return
	}

	c.Status(http.StatusNoContent)
}

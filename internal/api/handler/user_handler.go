package handler

import (
	"net/http"
	"strings"

	"acl-challenge/internal/usecase"

	"github.com/gin-gonic/gin"
)

type updateUserRequest struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func UpdateUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		respondWithError(c, usecase.ErrInvalidInput, "update user id is required")
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, usecase.ErrInvalidInput, "update user request body invalid")
		return
	}

	if strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.PasswordHash) == "" {
		respondWithError(c, usecase.ErrInvalidInput, "update user has no fields to update")
		return
	}

	Success(c, http.StatusOK, gin.H{"message": "stub: updated"})
}

func DeleteUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		respondWithError(c, usecase.ErrInvalidInput, "delete user id is required")
		return
	}

	c.Status(http.StatusNoContent)
}

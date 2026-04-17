package handler

import (
	"net/http"
	"strings"

	"acl-challenge/pkg/response"

	"github.com/gin-gonic/gin"
)

type updateUserRequest struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func UpdateUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.PasswordHash) == "" {
		response.Error(c, http.StatusBadRequest, "at least one field is required")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "stub: updated"})
}

func DeleteUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	c.Status(http.StatusNoContent)
}

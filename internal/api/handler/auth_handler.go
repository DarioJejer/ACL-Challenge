package handler

import (
	"net/http"
	"strings"

	"acl-challenge/pkg/response"
	"github.com/gin-gonic/gin"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		response.Error(c, http.StatusBadRequest, "email and password are required")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"message": "stub: user registered"})
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		response.Error(c, http.StatusBadRequest, "email and password are required")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"token": "stub-token"})
}

package handler

import (
	"net/http"
	"strings"

	"acl-challenge/internal/usecase"

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
		logAndRespondWithError(c, usecase.ErrInvalidInput, "register request body invalid")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "register missing required fields")
		return
	}

	Success(c, http.StatusCreated, gin.H{"message": "stub: user registered"})
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "login request body invalid")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "login missing required fields")
		return
	}

	Success(c, http.StatusOK, gin.H{"token": "stub-token"})
}

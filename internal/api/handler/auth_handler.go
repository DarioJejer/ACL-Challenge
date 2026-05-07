package handler

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"acl-challenge/internal/usecase"

	"github.com/google/uuid"
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

type AuthHandler struct {
	userUC *usecase.UserUseCase
}

func NewAuthHandler(userUC *usecase.UserUseCase) *AuthHandler {
	return &AuthHandler{userUC: userUC}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "register request body invalid")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "register missing required fields")
		return
	}

	user, err := h.userUC.Register(c.Request.Context(), usecase.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		logAndRespondWithError(c, err, "register failed")
		return
	}

	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = "dev-secret"
	}

	token, err := GenerateToken(user.ID, secret)
	if err != nil {
		logAndRespondWithError(c, usecase.ErrInternalServer, "register token generation failed")
		return
	}

	SetAuthCookie(c, token)
	Success(c, http.StatusCreated, toUserDTO(user))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "login request body invalid")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "login missing required fields")
		return
	}

	Success(c, http.StatusOK, gin.H{"token": "token"})
}

// TODO(issue-25): replace with signed JWT implementation.
func GenerateToken(userID uuid.UUID, secret string) (string, error) {
	if userID == uuid.Nil || strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("invalid token input")
	}
	return fmt.Sprintf("stub-token-%s", userID.String()), nil
}

// TODO(issue-25): adjust cookie flags/options when JWT auth is fully implemented.
func SetAuthCookie(c *gin.Context, token string) {
	c.SetCookie("auth_token", token, 3600, "/", "", false, true)
}

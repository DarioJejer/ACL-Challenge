package handler

import (
	"net/http"
	"os"
	"strings"

	"acl-challenge/internal/api/middleware"
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

	if err := issueAuthCookie(c, user.ID.String()); err != nil {
		logAndRespondWithError(c, usecase.ErrInternalServer, "register token generation failed")
		return
	}

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

	user, err := h.userUC.Login(c.Request.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		logAndRespondWithError(c, err, "login failed")
		return
	}

	if err := issueAuthCookie(c, user.ID.String()); err != nil {
		logAndRespondWithError(c, usecase.ErrInternalServer, "login token generation failed")
		return
	}

	Success(c, http.StatusOK, toUserDTO(user))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	middleware.ClearAuthCookie(c)
	Success(c, http.StatusOK, nil)
}

// issueAuthCookie generates a JWT for userID and writes it to the auth cookie.
// Centralised so register and login share the exact same logic.
func issueAuthCookie(c *gin.Context, userID string) error {
	token, err := middleware.GenerateToken(userID, jwtSecret())
	if err != nil {
		return err
	}
	middleware.SetAuthCookie(c, token)
	return nil
}

func jwtSecret() string {
	if s := strings.TrimSpace(os.Getenv("JWT_SECRET")); s != "" {
		return s
	}
	return "dev-secret"
}

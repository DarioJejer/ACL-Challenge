package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthCookieName  = "auth_token"
	cookieMaxAgeSec = 3600
)

// Claims is the application's JWT claim set.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken signs a new HS256 JWT for the given user.
// Tokens expire 1 hour after issuance.
func GenerateToken(userID string, secret string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("middleware: jwt: empty user id")
	}
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("middleware: jwt: empty secret")
	}

	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// SetAuthCookie writes the auth cookie with the project's hardening defaults.
// `Secure` is enabled only when APP_ENV=production so the cookie still works
// over plain HTTP in local development.
func SetAuthCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   cookieMaxAgeSec,
		HttpOnly: true,
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearAuthCookie instructs the browser to delete the auth cookie immediately.
func ClearAuthCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
	})
}

func isProduction() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
}

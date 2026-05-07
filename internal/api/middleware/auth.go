package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	contextUserIDKey = "userID"
)

// ValidateToken returns a Gin middleware that validates a JWT supplied in the
// auth cookie, injects `userID` into the context, and refreshes the cookie on
// every successful validation (sliding expiration window). It is stateless: no
// database lookup is performed.
//
// Failure modes (each aborts the request):
//   - Missing/empty cookie  -> 401 "authentication required"
//   - Parse / signature err -> 401 "invalid token"
//   - Expired token         -> 401 "token expired" (also clears the cookie)
func ValidateToken(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(AuthCookieName)
		if err != nil || strings.TrimSpace(raw) == "" {
			abortJSON(c, http.StatusUnauthorized, "authentication required")
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || token == nil || !token.Valid {
			if errors.Is(err, jwt.ErrTokenExpired) {
				ClearAuthCookie(c)
				abortJSON(c, http.StatusUnauthorized, "token expired")
				return
			}
			abortJSON(c, http.StatusUnauthorized, "invalid token")
			return
		}

		// Defensive expiration check: jwt/v5 already rejects expired tokens, but
		// we re-check in case the library skips it for any reason (e.g. custom
		// parser options in the future).
		if claims.ExpiresAt == nil || !claims.ExpiresAt.After(time.Now()) {
			ClearAuthCookie(c)
			abortJSON(c, http.StatusUnauthorized, "token expired")
			return
		}

		if strings.TrimSpace(claims.UserID) == "" {
			abortJSON(c, http.StatusUnauthorized, "invalid token")
			return
		}

		c.Set(contextUserIDKey, claims.UserID)

		newToken, err := GenerateToken(claims.UserID, secret)
		if err != nil {
			abortJSON(c, http.StatusUnauthorized, "invalid token")
			return
		}
		SetAuthCookie(c, newToken)

		c.Next()
	}
}

// GetUserIDFromContext returns the userID injected by ValidateToken.
// Returns ("", false) when the value is missing or not a string.
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(contextUserIDKey)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", false
	}
	return id, true
}

func abortJSON(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"code":    "UNAUTHORIZED",
		"message": message,
	})
}

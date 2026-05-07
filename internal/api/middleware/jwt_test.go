package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"acl-challenge/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	t.Run("happy path returns parseable HS256 token", func(t *testing.T) {
		t.Parallel()
		userID := uuid.NewString()
		secret := "test-secret"

		signed, err := middleware.GenerateToken(userID, secret)
		require.NoError(t, err)
		require.NotEmpty(t, signed)

		parsed, err := jwt.ParseWithClaims(signed, &middleware.Claims{}, func(_ *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, err)
		require.True(t, parsed.Valid)

		claims, ok := parsed.Claims.(*middleware.Claims)
		require.True(t, ok)
		require.Equal(t, userID, claims.UserID)
		require.NotNil(t, claims.ExpiresAt)
	})

	t.Run("rejects empty user id", func(t *testing.T) {
		t.Parallel()
		_, err := middleware.GenerateToken("", "secret")
		require.Error(t, err)
	})

	t.Run("rejects empty secret", func(t *testing.T) {
		t.Parallel()
		_, err := middleware.GenerateToken(uuid.NewString(), "")
		require.Error(t, err)
	})
}

func TestSetAuthCookie(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	middleware.SetAuthCookie(c, "abc")

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	require.Equal(t, middleware.AuthCookieName, cookie.Name)
	require.Equal(t, "abc", cookie.Value)
	require.Equal(t, "/", cookie.Path)
	require.Equal(t, 3600, cookie.MaxAge)
	require.True(t, cookie.HttpOnly)
	require.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
}

func TestClearAuthCookie(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	middleware.ClearAuthCookie(c)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	require.Equal(t, middleware.AuthCookieName, cookie.Name)
	require.Equal(t, "", cookie.Value)
	require.Equal(t, -1, cookie.MaxAge)
}

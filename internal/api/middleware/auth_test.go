package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"acl-challenge/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testSecret = "validate-token-secret"

func init() {
	gin.SetMode(gin.TestMode)
}

func newProtectedRouter(handler gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.GET("/protected", middleware.ValidateToken(testSecret), handler)
	return r
}

func performRequest(r http.Handler, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func cookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, ck := range cookies {
		if ck.Name == name {
			return ck
		}
	}
	return nil
}

func TestValidateToken_MissingCookie(t *testing.T) {
	t.Parallel()
	r := newProtectedRouter(func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := performRequest(r, nil)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "authentication required")
}

func TestValidateToken_EmptyCookieValue(t *testing.T) {
	t.Parallel()
	r := newProtectedRouter(func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := performRequest(r, &http.Cookie{Name: middleware.AuthCookieName, Value: ""})
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "authentication required")
}

func TestValidateToken_MalformedToken(t *testing.T) {
	t.Parallel()
	r := newProtectedRouter(func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := performRequest(r, &http.Cookie{Name: middleware.AuthCookieName, Value: "not-a-jwt"})
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "invalid token")
}

func TestValidateToken_WrongSecret(t *testing.T) {
	t.Parallel()
	r := newProtectedRouter(func(c *gin.Context) { c.Status(http.StatusOK) })

	other, err := middleware.GenerateToken(uuid.NewString(), "different-secret")
	require.NoError(t, err)

	rec := performRequest(r, &http.Cookie{Name: middleware.AuthCookieName, Value: other})
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "invalid token")
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	t.Parallel()
	r := newProtectedRouter(func(c *gin.Context) { c.Status(http.StatusOK) })

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.Claims{
		UserID: uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	signed, err := expired.SignedString([]byte(testSecret))
	require.NoError(t, err)

	rec := performRequest(r, &http.Cookie{Name: middleware.AuthCookieName, Value: signed})
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "token expired")

	cleared := cookieByName(rec.Result().Cookies(), middleware.AuthCookieName)
	require.NotNil(t, cleared, "middleware should clear the auth cookie when token is expired")
	require.Equal(t, -1, cleared.MaxAge)
	require.Equal(t, "", cleared.Value)
}

func TestValidateToken_HappyPathInjectsUserIDAndRefreshesCookie(t *testing.T) {
	t.Parallel()

	userID := uuid.NewString()
	var captured string
	r := newProtectedRouter(func(c *gin.Context) {
		id, ok := middleware.GetUserIDFromContext(c)
		require.True(t, ok)
		captured = id
		c.Status(http.StatusOK)
	})

	signed, err := middleware.GenerateToken(userID, testSecret)
	require.NoError(t, err)

	rec := performRequest(r, &http.Cookie{Name: middleware.AuthCookieName, Value: signed})
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, userID, captured)

	refreshed := cookieByName(rec.Result().Cookies(), middleware.AuthCookieName)
	require.NotNil(t, refreshed, "middleware should refresh the auth cookie on success")
	require.NotEmpty(t, refreshed.Value)
	require.Equal(t, 3600, refreshed.MaxAge)
}

func TestGetUserIDFromContext(t *testing.T) {
	t.Parallel()

	t.Run("returns false when not set", func(t *testing.T) {
		t.Parallel()
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		id, ok := middleware.GetUserIDFromContext(c)
		require.False(t, ok)
		require.Equal(t, "", id)
	})

	t.Run("returns false for non-string value", func(t *testing.T) {
		t.Parallel()
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userID", 42)
		id, ok := middleware.GetUserIDFromContext(c)
		require.False(t, ok)
		require.Equal(t, "", id)
	})

	t.Run("returns the injected userID", func(t *testing.T) {
		t.Parallel()
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userID", "abc-123")
		id, ok := middleware.GetUserIDFromContext(c)
		require.True(t, ok)
		require.Equal(t, "abc-123", id)
	})
}

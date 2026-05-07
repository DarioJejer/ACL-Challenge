package testhelper

import (
	"net/http"
	"testing"

	"acl-challenge/internal/api/middleware"

	"github.com/stretchr/testify/require"
)

// AuthCookieFor mints a valid auth cookie for the given userID using the
// project's JWT secret resolution (env-driven, with dev fallback). Tests can
// attach the returned cookie to requests targeting protected routes without
// going through the register/login HTTP flow.
func AuthCookieFor(t *testing.T, userID string) *http.Cookie {
	t.Helper()

	token, err := middleware.GenerateToken(userID, middleware.JWTSecret())
	require.NoError(t, err)

	return &http.Cookie{
		Name:  middleware.AuthCookieName,
		Value: token,
		Path:  "/",
	}
}

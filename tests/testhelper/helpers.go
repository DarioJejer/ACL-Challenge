package testhelper

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func AssertErrorIs(t *testing.T, err error, target error) {
	t.Helper()
	require.ErrorIs(t, err, target)
}

func NewContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

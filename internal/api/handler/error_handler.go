package handler

import (
	"errors"
	"log/slog"

	"acl-challenge/internal/usecase"

	"github.com/gin-gonic/gin"
)

func respondWithError(c *gin.Context, err error, message string) {
	logAttrs := []any{
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"error", err.Error(),
	}

	if errors.Is(err, usecase.ErrDatabase) || errors.Is(err, usecase.ErrInternalServer) {
		slog.Error(message, logAttrs...)
	} else {
		slog.Warn(message, logAttrs...)
	}

	ErrorResponse(c, err)
}

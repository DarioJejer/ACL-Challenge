package handler

import (
	"errors"
	"net/http"

	"acl-challenge/internal/usecase"

	"github.com/gin-gonic/gin"
)

func HTTPStatusFromError(err error) int {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, usecase.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, usecase.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, usecase.ErrUnsupportedChannel):
		return http.StatusUnprocessableEntity
	case errors.Is(err, usecase.ErrDatabase):
		return http.StatusServiceUnavailable
	case errors.Is(err, usecase.ErrInternalServer):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func ErrorResponse(c *gin.Context, err error) {
	statusCode := HTTPStatusFromError(err)
	Error(c, statusCode, err.Error())
}

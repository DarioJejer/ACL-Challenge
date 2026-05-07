package handler

import (
	"errors"
	"net/http"

	"acl-challenge/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponseCode string

const (
	NotFound            ErrorResponseCode = "NOT_FOUND"
	Conflict            ErrorResponseCode = "CONFLICT"
	InvalidInput        ErrorResponseCode = "INVALID_INPUT"
	UnsupportedChannel  ErrorResponseCode = "UNSUPPORTED_CHANNEL"
	Unauthorized        ErrorResponseCode = "UNAUTHORIZED"
	DatabaseError       ErrorResponseCode = "DATABASE_ERROR"
	InternalServerError ErrorResponseCode = "INTERNAL_SERVER_ERROR"
)

func HTTPResponseFromError(err error) (int, Envelope) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		return http.StatusNotFound, Envelope{Success: false, Code: string(NotFound), Message: usecase.ErrNotFound.Error()}
	case errors.Is(err, usecase.ErrConflict):
		return http.StatusConflict, Envelope{Success: false, Code: string(Conflict), Message: usecase.ErrConflict.Error()}
	case errors.Is(err, usecase.ErrInvalidInput):
		return http.StatusBadRequest, Envelope{Success: false, Code: string(InvalidInput), Message: usecase.ErrInvalidInput.Error()}
	case errors.Is(err, usecase.ErrUnsupportedChannel):
		return http.StatusUnprocessableEntity, Envelope{Success: false, Code: string(UnsupportedChannel), Message: usecase.ErrUnsupportedChannel.Error()}
	case errors.Is(err, usecase.ErrUnauthorized):
		return http.StatusUnauthorized, Envelope{Success: false, Code: string(Unauthorized), Message: "invalid credentials"}
	case errors.Is(err, usecase.ErrDatabase):
		return http.StatusInternalServerError, Envelope{Success: false, Code: string(DatabaseError), Message: usecase.ErrDatabase.Error()}
	case errors.Is(err, usecase.ErrInternalServer):
		return http.StatusInternalServerError, Envelope{Success: false, Code: string(InternalServerError), Message: usecase.ErrInternalServer.Error()}
	default:
		return http.StatusInternalServerError, Envelope{Success: false, Code: string(InternalServerError), Message: err.Error()}
	}
}

func ErrorResponse(c *gin.Context, err error) {
	statusCode, envelope := HTTPResponseFromError(err)
	c.JSON(statusCode, envelope)
}

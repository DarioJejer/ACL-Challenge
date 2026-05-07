package usecase

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrConflict           = errors.New("resource already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnsupportedChannel = errors.New("unsupported notification channel")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("access denied")
	ErrDatabase           = errors.New("database error")
	ErrInternalServer     = errors.New("internal server error")
)

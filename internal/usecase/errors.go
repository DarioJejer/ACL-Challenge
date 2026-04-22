package usecase

import "errors"

var (
	ErrNotFound = errors.New("resource not found")
	ErrDatabase = errors.New("database error")
)

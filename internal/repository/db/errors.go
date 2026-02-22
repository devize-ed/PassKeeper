package db

import (
	"errors"
)

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrItemNotFound      = errors.New("item not found")
	ErrWrongUserID       = errors.New("Item does not belong to the user")
	ErrTimestampTooOld   = errors.New("timestamp is older than the updated at")
	ErrItemAlreadyExists = errors.New("item already exists")
)

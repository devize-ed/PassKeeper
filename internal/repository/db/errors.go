package db

import (
	"errors"
)

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

// Sentinel errors returned by the database layer.
var (
	ErrUserAlreadyExists = errors.New("user already exists")                    // ErrUserAlreadyExists indicates duplicate username.
	ErrUserNotFound      = errors.New("user not found")                         // ErrUserNotFound indicates no user with given username.
	ErrInvalidPassword   = errors.New("invalid password")                       // ErrInvalidPassword indicates password mismatch.
	ErrItemNotFound      = errors.New("item not found")                         // ErrItemNotFound indicates no item with given ID.
	ErrTimestampTooOld   = errors.New("timestamp is older than the updated at") // ErrTimestampTooOld indicates conflict.
	ErrItemAlreadyExists = errors.New("item already exists")                    // ErrItemAlreadyExists indicates duplicate item.
)

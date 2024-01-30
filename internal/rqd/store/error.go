package store

import "errors"

var (
	ErrKeyAlreadyExists = errors.New("key already exists")
	ErrKeyNotFound      = errors.New("key not found")
)

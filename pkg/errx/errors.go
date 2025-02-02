package errx

import (
	"errors"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrDuplicateKey = errors.New("duplicate key")
)

type HTTPError struct {
	Message string `json:"message"`
}

func (e HTTPError) Error() string {
	return e.Message
}

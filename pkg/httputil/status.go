package httputil

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	StatusCode int
	Code       int
	Message    string
}

func (e *Error) Error() string {
	switch {
	case e == nil:
		return "<nil>"
	case e.Message == "":
		return http.StatusText(e.StatusCode)
	default:
		return e.Message
	}
}

func NewStatus(statusCode, code int, message string) error {
	return &Error{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

func Errorf(statusCode, code int, format string, a ...any) error {
	return &Error{
		StatusCode: statusCode,
		Code:       code,
		Message:    fmt.Sprintf(format, a...),
	}
}

func StatusWrap(statusCode, code int, err error) error {
	return &Error{
		StatusCode: statusCode,
		Code:       code,
		Message:    err.Error(),
	}
}

func StatusWrapE(statusCode, code int, message string, err error) error {
	return &Error{
		StatusCode: statusCode,
		Code:       code,
		Message:    fmt.Sprintf("%s: %s", message, err.Error()),
	}
}

func As(err error) (*Error, bool) {
	var statusErr *Error
	return statusErr, errors.As(err, &statusErr)
}

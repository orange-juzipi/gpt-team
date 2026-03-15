package apperr

import (
	"errors"
	"net/http"
)

// Error carries an HTTP status and a stable error code that handlers can expose.
type Error struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func New(status int, code, message string) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func Wrap(status int, code, message string, cause error) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func BadRequest(code, message string) *Error {
	return New(http.StatusBadRequest, code, message)
}

func Unauthorized(code, message string) *Error {
	return New(http.StatusUnauthorized, code, message)
}

func Forbidden(code, message string) *Error {
	return New(http.StatusForbidden, code, message)
}

func NotFound(code, message string) *Error {
	return New(http.StatusNotFound, code, message)
}

func Conflict(code, message string) *Error {
	return New(http.StatusConflict, code, message)
}

func Upstream(code, message string, cause error) *Error {
	return Wrap(http.StatusBadGateway, code, message, cause)
}

func Internal(code, message string, cause error) *Error {
	return Wrap(http.StatusInternalServerError, code, message, cause)
}

func Status(err error) int {
	var appError *Error
	if errors.As(err, &appError) {
		return appError.Status
	}

	return http.StatusInternalServerError
}

func Code(err error) string {
	var appError *Error
	if errors.As(err, &appError) {
		return appError.Code
	}

	return "internal_error"
}

func Message(err error) string {
	var appError *Error
	if errors.As(err, &appError) {
		return appError.Message
	}

	return "internal server error"
}

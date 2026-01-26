// Package apperror provides domain error types with HTTP status mapping.
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Code represents an application error code.
type Code string

// Error codes for common application errors.
const (
	CodeNotFound       Code = "NOT_FOUND"
	CodeAlreadyExists  Code = "ALREADY_EXISTS"
	CodeInvalidInput   Code = "INVALID_INPUT"
	CodeUnauthorized   Code = "UNAUTHORIZED"
	CodeInternal       Code = "INTERNAL_ERROR"
	CodeHashingFailed  Code = "HASHING_FAILED"
	CodeTokenGenFailed Code = "TOKEN_GENERATION_FAILED"
)

// Error is the application's standard error type.
type Error struct {
	Code    Code   // Machine-readable code
	Message string // Human-readable message
	Err     error  // Wrapped underlying error (for logging, not exposed to API)
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the appropriate HTTP status code for the error.
func (e *Error) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists:
		return http.StatusConflict
	case CodeInvalidInput:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeInternal, CodeHashingFailed, CodeTokenGenFailed:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new Error.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap creates a new Error wrapping an underlying error.
func Wrap(code Code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

// Predefined sentinel errors for common cases.
var (
	ErrNotFound           = New(CodeNotFound, "resource not found")
	ErrUserNotFound       = New(CodeNotFound, "user not found")
	ErrUserAlreadyExists  = New(CodeAlreadyExists, "user already exists")
	ErrInvalidInput       = New(CodeInvalidInput, "invalid input")
	ErrInvalidCredentials = New(CodeUnauthorized, "invalid credentials")
	ErrUnauthorized       = New(CodeUnauthorized, "unauthorized")
)

// ErrorResponse represents an API error response for Swagger documentation.
type ErrorResponse struct {
	Error string `json:"error" example:"user not found"`
	Code  Code   `json:"code" example:"NOT_FOUND"`
}

// Is checks if the target error has the same Code.
func Is(err error, code Code) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// AsAppError attempts to extract an *Error from err.
func AsAppError(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

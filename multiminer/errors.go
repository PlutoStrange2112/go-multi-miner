package multiminer

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes for structured error handling
const (
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeNotImplemented   = "NOT_IMPLEMENTED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeConnectionFailed = "CONNECTION_FAILED"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeDriverNotFound   = "DRIVER_NOT_FOUND"
	ErrCodeDeviceError      = "DEVICE_ERROR"
	ErrCodeInternalError    = "INTERNAL_ERROR"
)

// MultiMinerError provides structured error information
type MultiMinerError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Cause   error  `json:"-"`
}

func (e *MultiMinerError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *MultiMinerError) Unwrap() error {
	return e.Cause
}

// HTTPStatus returns appropriate HTTP status code for the error
func (e *MultiMinerError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeNotFound, ErrCodeDriverNotFound:
		return http.StatusNotFound
	case ErrCodeNotImplemented:
		return http.StatusNotImplemented
	case ErrCodeInvalidInput:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeConnectionFailed, ErrCodeTimeout, ErrCodeDeviceError:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// Predefined errors
var (
	ErrNotFound       = &MultiMinerError{Code: ErrCodeNotFound, Message: "device not found"}
	ErrNotImplemented = &MultiMinerError{Code: ErrCodeNotImplemented, Message: "not implemented"}
)

// Error constructors
func NewNotFoundError(message string) *MultiMinerError {
	return &MultiMinerError{Code: ErrCodeNotFound, Message: message}
}

func NewInvalidInputError(message string) *MultiMinerError {
	return &MultiMinerError{Code: ErrCodeInvalidInput, Message: message}
}

func NewConnectionError(details string, cause error) *MultiMinerError {
	return &MultiMinerError{
		Code:    ErrCodeConnectionFailed,
		Message: "failed to connect to device",
		Details: details,
		Cause:   cause,
	}
}

func NewTimeoutError(details string) *MultiMinerError {
	return &MultiMinerError{
		Code:    ErrCodeTimeout,
		Message: "operation timed out",
		Details: details,
	}
}

func NewDriverNotFoundError() *MultiMinerError {
	return &MultiMinerError{Code: ErrCodeDriverNotFound, Message: "no suitable driver found"}
}

func NewDeviceError(message, details string, cause error) *MultiMinerError {
	return &MultiMinerError{
		Code:    ErrCodeDeviceError,
		Message: message,
		Details: details,
		Cause:   cause,
	}
}

// IsMultiMinerError checks if an error is a MultiMinerError
func IsMultiMinerError(err error) (*MultiMinerError, bool) {
	var mErr *MultiMinerError
	if errors.As(err, &mErr) {
		return mErr, true
	}
	return nil, false
}

// WrapError wraps a generic error with MultiMinerError
func WrapError(err error, code, message string) *MultiMinerError {
	return &MultiMinerError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}
// Package db provides database connection management and error handling utilities.
// It includes structured error types and connection helpers for the SCG database toolkit.
package db

import (
	"errors"
	"fmt"
)

var (
	// ErrRecordNotFound indicates that a requested database record was not found.
	ErrRecordNotFound = errors.New("record not found")
	// ErrConfigValidation indicates that database configuration validation failed.
	ErrConfigValidation = errors.New("config validation failed")
	// ErrInvalidAdapter indicates that an invalid adapter type was provided.
	ErrInvalidAdapter = errors.New("invalid adapter type")
	// ErrAdapterLookup indicates that adapter lookup failed.
	ErrAdapterLookup = errors.New("adapter lookup failed")
	// ErrAdapterConnect indicates that adapter connection failed.
	ErrAdapterConnect = errors.New("adapter connect failed")
	// ErrConnectionPing indicates that database connection ping failed.
	ErrConnectionPing = errors.New("connection ping failed")
)

// Error represents a structured database error with context
type (
	Error struct {
		Operation string
		Message   string
		Err       error
	}
)

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("db operation '%s' failed: %s: %v", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("db operation '%s' failed: %s", e.Operation, e.Message)
}

// Unwrap returns the underlying error for error wrapping
func (e *Error) Unwrap() error {
	return e.Err
}

// Is implements error matching for errors.Is
func (e *Error) Is(target error) bool {
	if dbErr, ok := target.(*Error); ok {
		return e.Operation == dbErr.Operation && e.Message == dbErr.Message
	}
	return errors.Is(e.Err, target)
}

// NewError creates a new structured database error
func NewError(operation, message string, err error) *Error {
	return &Error{
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// NewConfigValidationError creates a new Error for configuration validation failures.
func NewConfigValidationError(err error) error {
	return NewError("Connect", "config validation failed", err)
}

// NewInvalidAdapterError creates a new Error for invalid adapter type errors.
func NewInvalidAdapterError() error {
	return NewError("Connect", "invalid adapter type provided in config", ErrInvalidAdapter)
}

// NewAdapterLookupError creates a new Error for adapter lookup failures.
func NewAdapterLookupError(err error) error {
	return NewError("Connect", "adapter lookup failed", err)
}

// NewAdapterConnectError creates a new Error for adapter connection failures.
func NewAdapterConnectError(err error) error {
	return NewError("Connect", "adapter connect failed", err)
}

// NewConnectionPingError creates a new Error for database ping failures.
func NewConnectionPingError(err error) error {
	return NewError("Connect", "initial database ping failed", err)
}

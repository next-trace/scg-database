package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Unwrap(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected error
	}{
		{
			name: "error with underlying error",
			err: &Error{
				Operation: "test",
				Message:   "test message",
				Err:       errors.New("underlying error"),
			},
			expected: errors.New("underlying error"),
		},
		{
			name: "error without underlying error",
			err: &Error{
				Operation: "test",
				Message:   "test message",
				Err:       nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Unwrap()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.Error(), result.Error())
			}
		})
	}
}

func TestNewInvalidAdapterError(t *testing.T) {
	err := NewInvalidAdapterError()

	assert.NotNil(t, err)
	assert.IsType(t, &Error{}, err)

	dbErr := err.(*Error)
	assert.Equal(t, "Connect", dbErr.Operation)
	assert.Equal(t, "invalid adapter type provided in config", dbErr.Message)
	assert.Equal(t, ErrInvalidAdapter, dbErr.Err)

	// Test error message format
	expectedMsg := "db operation 'Connect' failed: invalid adapter type provided in config: invalid adapter type"
	assert.Equal(t, expectedMsg, err.Error())
}

func TestError_Is_EdgeCases(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := &Error{
		Operation: "test",
		Message:   "test message",
		Err:       baseErr,
	}

	tests := []struct {
		name     string
		err      *Error
		target   error
		expected bool
	}{
		{
			name:     "matching Error struct",
			err:      wrappedErr,
			target:   &Error{Operation: "test", Message: "test message"},
			expected: true,
		},
		{
			name:     "non-matching Error struct - different operation",
			err:      wrappedErr,
			target:   &Error{Operation: "different", Message: "test message"},
			expected: false,
		},
		{
			name:     "non-matching Error struct - different message",
			err:      wrappedErr,
			target:   &Error{Operation: "test", Message: "different message"},
			expected: false,
		},
		{
			name:     "matching underlying error",
			err:      wrappedErr,
			target:   baseErr,
			expected: true,
		},
		{
			name:     "non-matching underlying error",
			err:      wrappedErr,
			target:   errors.New("different error"),
			expected: false,
		},
		{
			name:     "non-Error target",
			err:      wrappedErr,
			target:   errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Is(tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestError_Error_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error with underlying error",
			err: &Error{
				Operation: "Connect",
				Message:   "connection failed",
				Err:       errors.New("network timeout"),
			},
			expected: "db operation 'Connect' failed: connection failed: network timeout",
		},
		{
			name: "error without underlying error",
			err: &Error{
				Operation: "Query",
				Message:   "invalid query",
				Err:       nil,
			},
			expected: "db operation 'Query' failed: invalid query",
		},
		{
			name: "error with empty operation",
			err: &Error{
				Operation: "",
				Message:   "some error",
				Err:       nil,
			},
			expected: "db operation '' failed: some error",
		},
		{
			name: "error with empty message",
			err: &Error{
				Operation: "Test",
				Message:   "",
				Err:       nil,
			},
			expected: "db operation 'Test' failed: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewError(t *testing.T) {
	underlyingErr := errors.New("underlying error")

	tests := []struct {
		name      string
		operation string
		message   string
		err       error
	}{
		{
			name:      "complete error",
			operation: "TestOp",
			message:   "test message",
			err:       underlyingErr,
		},
		{
			name:      "error without underlying error",
			operation: "TestOp",
			message:   "test message",
			err:       nil,
		},
		{
			name:      "empty operation",
			operation: "",
			message:   "test message",
			err:       underlyingErr,
		},
		{
			name:      "empty message",
			operation: "TestOp",
			message:   "",
			err:       underlyingErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewError(tt.operation, tt.message, tt.err)

			assert.NotNil(t, result)
			assert.Equal(t, tt.operation, result.Operation)
			assert.Equal(t, tt.message, result.Message)
			assert.Equal(t, tt.err, result.Err)
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error constants are properly defined
	assert.NotNil(t, ErrRecordNotFound)
	assert.NotNil(t, ErrConfigValidation)
	assert.NotNil(t, ErrInvalidAdapter)
	assert.NotNil(t, ErrAdapterLookup)
	assert.NotNil(t, ErrAdapterConnect)
	assert.NotNil(t, ErrConnectionPing)

	// Test error messages
	assert.Equal(t, "record not found", ErrRecordNotFound.Error())
	assert.Equal(t, "config validation failed", ErrConfigValidation.Error())
	assert.Equal(t, "invalid adapter type", ErrInvalidAdapter.Error())
	assert.Equal(t, "adapter lookup failed", ErrAdapterLookup.Error())
	assert.Equal(t, "adapter connect failed", ErrAdapterConnect.Error())
	assert.Equal(t, "connection ping failed", ErrConnectionPing.Error())
}

func TestNewConfigValidationError(t *testing.T) {
	underlyingErr := errors.New("validation failed")
	err := NewConfigValidationError(underlyingErr)

	assert.NotNil(t, err)
	assert.IsType(t, &Error{}, err)

	dbErr := err.(*Error)
	assert.Equal(t, "Connect", dbErr.Operation)
	assert.Equal(t, "config validation failed", dbErr.Message)
	assert.Equal(t, underlyingErr, dbErr.Err)
}

func TestNewAdapterLookupError(t *testing.T) {
	underlyingErr := errors.New("adapter not found")
	err := NewAdapterLookupError(underlyingErr)

	assert.NotNil(t, err)
	assert.IsType(t, &Error{}, err)

	dbErr := err.(*Error)
	assert.Equal(t, "Connect", dbErr.Operation)
	assert.Equal(t, "adapter lookup failed", dbErr.Message)
	assert.Equal(t, underlyingErr, dbErr.Err)
}

func TestNewAdapterConnectError(t *testing.T) {
	underlyingErr := errors.New("connection failed")
	err := NewAdapterConnectError(underlyingErr)

	assert.NotNil(t, err)
	assert.IsType(t, &Error{}, err)

	dbErr := err.(*Error)
	assert.Equal(t, "Connect", dbErr.Operation)
	assert.Equal(t, "adapter connect failed", dbErr.Message)
	assert.Equal(t, underlyingErr, dbErr.Err)
}

func TestNewConnectionPingError(t *testing.T) {
	underlyingErr := errors.New("ping failed")
	err := NewConnectionPingError(underlyingErr)

	assert.NotNil(t, err)
	assert.IsType(t, &Error{}, err)

	dbErr := err.(*Error)
	assert.Equal(t, "Connect", dbErr.Operation)
	assert.Equal(t, "initial database ping failed", dbErr.Message)
	assert.Equal(t, underlyingErr, dbErr.Err)
}

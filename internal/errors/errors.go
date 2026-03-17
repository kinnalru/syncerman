package errors

import "fmt"

// ErrorType represents different categories of errors in the Syncerman application.
type ErrorType int

const (
	// TypeConfig indicates a configuration-related error.
	TypeConfig ErrorType = iota
	// TypeRclone indicates an rclone execution error.
	TypeRclone
	// TypeValidation indicates a validation error.
	TypeValidation
)

// String returns the string representation of the ErrorType.
func (e ErrorType) String() string {
	switch e {
	case TypeConfig:
		return "CONFIG"
	case TypeRclone:
		return "RCLONE"
	case TypeValidation:
		return "VALIDATION"
	default:
		return "UNKNOWN"
	}
}

// SyncermanError represents a custom error type with typed error categories.
type SyncermanError struct {
	// Type indicates the category of the error.
	Type ErrorType
	// Message provides a human-readable description of the error.
	Message string
	// Err holds the underlying error, if any.
	Err error
}

// Error returns the error message in the format "TYPE: message" or "TYPE: message: underlying error".
func (e *SyncermanError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *SyncermanError) Unwrap() error {
	return e.Err
}

func newSyncermanError(errorType ErrorType, message string, err error) *SyncermanError {
	return &SyncermanError{
		Type:    errorType,
		Message: message,
		Err:     err,
	}
}

// NewConfigError creates a new SyncermanError of type TypeConfig.
func NewConfigError(message string, err error) *SyncermanError {
	return newSyncermanError(TypeConfig, message, err)
}

// NewRcloneError creates a new SyncermanError of type TypeRclone.
func NewRcloneError(message string, err error) *SyncermanError {
	return newSyncermanError(TypeRclone, message, err)
}

// NewValidationError creates a new SyncermanError of type TypeValidation.
func NewValidationError(message string, err error) *SyncermanError {
	return newSyncermanError(TypeValidation, message, err)
}

func isErrorType(err error, errorType ErrorType) bool {
	if syncErr, ok := err.(*SyncermanError); ok {
		return syncErr.Type == errorType
	}
	return false
}

// IsConfigError reports whether err is a TypeConfig error.
func IsConfigError(err error) bool {
	return isErrorType(err, TypeConfig)
}

// IsRcloneError reports whether err is a TypeRclone error.
func IsRcloneError(err error) bool {
	return isErrorType(err, TypeRclone)
}

// IsValidationError reports whether err is a TypeValidation error.
func IsValidationError(err error) bool {
	return isErrorType(err, TypeValidation)
}

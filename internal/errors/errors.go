package errors

import "fmt"

type ErrorType int

const (
	TypeConfig ErrorType = iota
	TypeRclone
	TypeValidation
)

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

type SyncermanError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *SyncermanError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *SyncermanError) Unwrap() error {
	return e.Err
}

func NewConfigError(message string, err error) *SyncermanError {
	return &SyncermanError{
		Type:    TypeConfig,
		Message: message,
		Err:     err,
	}
}

func NewRcloneError(message string, err error) *SyncermanError {
	return &SyncermanError{
		Type:    TypeRclone,
		Message: message,
		Err:     err,
	}
}

func NewValidationError(message string, err error) *SyncermanError {
	return &SyncermanError{
		Type:    TypeValidation,
		Message: message,
		Err:     err,
	}
}

func IsConfigError(err error) bool {
	if syncErr, ok := err.(*SyncermanError); ok {
		return syncErr.Type == TypeConfig
	}
	return false
}

func IsRcloneError(err error) bool {
	if syncErr, ok := err.(*SyncermanError); ok {
		return syncErr.Type == TypeRclone
	}
	return false
}

func IsValidationError(err error) bool {
	if syncErr, ok := err.(*SyncermanError); ok {
		return syncErr.Type == TypeValidation
	}
	return false
}

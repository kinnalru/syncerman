package errors

import (
	"errors"
	"testing"
)

func TestNewConfigError(t *testing.T) {
	err := NewConfigError("test message", nil)
	if err.Type != TypeConfig {
		t.Errorf("expected TypeConfig, got %v", err.Type)
	}
	if err.Message != "test message" {
		t.Errorf("expected 'test message', got %v", err.Message)
	}
}

func TestNewConfigErrorWithWrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewConfigError("test message", baseErr)
	if err.Err != baseErr {
		t.Errorf("expected base error, got %v", err.Err)
	}
}

func TestNewRcloneError(t *testing.T) {
	err := NewRcloneError("test message", nil)
	if err.Type != TypeRclone {
		t.Errorf("expected TypeRclone, got %v", err.Type)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test message", nil)
	if err.Type != TypeValidation {
		t.Errorf("expected TypeValidation, got %v", err.Type)
	}
}

func TestSyncermanError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *SyncermanError
		contains []string
	}{
		{
			name:     "without wrap",
			err:      NewConfigError("test message", nil),
			contains: []string{"CONFIG", "test message"},
		},
		{
			name:     "with wrap",
			err:      NewConfigError("test message", errors.New("base error")),
			contains: []string{"CONFIG", "test message", "base error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, cont := range tt.contains {
				if !containsString(errStr, cont) {
					t.Errorf("expected %q in error string, got %q", cont, errStr)
				}
			}
		})
	}
}

func TestSyncermanError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewConfigError("test message", baseErr)

	if err.Unwrap() != baseErr {
		t.Errorf("expected base error from unwrap, got %v", err.Unwrap())
	}
}

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		etype    ErrorType
		expected string
	}{
		{TypeConfig, "CONFIG"},
		{TypeRclone, "RCLONE"},
		{TypeValidation, "VALIDATION"},
	}

	for _, tt := range tests {
		if tt.etype.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.etype.String())
		}
	}
}

func TestIsConfigError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"is config error", NewConfigError("test", nil), true},
		{"is rclone error", NewRcloneError("test", nil), false},
		{"is validation error", NewValidationError("test", nil), false},
		{"is generic error", errors.New("generic"), false},
		{"is nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConfigError(tt.err); got != tt.want {
				t.Errorf("IsConfigError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRcloneError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"is config error", NewConfigError("test", nil), false},
		{"is rclone error", NewRcloneError("test", nil), true},
		{"is validation error", NewValidationError("test", nil), false},
		{"is generic error", errors.New("generic"), false},
		{"is nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRcloneError(tt.err); got != tt.want {
				t.Errorf("IsRcloneError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"is config error", NewConfigError("test", nil), false},
		{"is rclone error", NewRcloneError("test", nil), false},
		{"is validation error", NewValidationError("test", nil), true},
		{"is generic error", errors.New("generic"), false},
		{"is nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidationError(tt.err); got != tt.want {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorUnwrapInterface(t *testing.T) {
	baseErr := errors.New("underlying")
	err := NewConfigError("wrapper", baseErr)

	if err.Unwrap() != baseErr {
		t.Error("Unwrap should return the base error")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && (containsString(s[1:], substr))
}

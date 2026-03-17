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
		{ErrorType(999), "UNKNOWN"},
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

func TestErrorChainWithErrorsIs(t *testing.T) {
	baseErr := errors.New("underlying error")
	wrappedErr := NewConfigError("wrapper message", baseErr)

	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should find the underlying error in the chain")
	}

	if errors.Is(wrappedErr, errors.New("different error")) {
		t.Error("errors.Is should not find a different error in the chain")
	}
}

func TestErrorChainWithErrorsAs(t *testing.T) {
	baseErr := errors.New("underlying error")
	wrappedErr := NewConfigError("wrapper message", baseErr)

	var syncErr *SyncermanError
	if !errors.As(wrappedErr, &syncErr) {
		t.Error("errors.As should find SyncermanError in the chain")
	}

	if syncErr.Type != TypeConfig {
		t.Errorf("expected TypeConfig, got %v", syncErr.Type)
	}

	if syncErr.Message != "wrapper message" {
		t.Errorf("expected 'wrapper message', got %v", syncErr.Message)
	}
}

func TestMultiLevelErrorWrapping(t *testing.T) {
	baseErr := errors.New("base error")
	level1 := NewConfigError("config problem", baseErr)

	var syncErr *SyncermanError
	if !errors.As(level1, &syncErr) {
		t.Error("errors.As should find SyncermanError at first level")
	}

	if !errors.Is(level1, baseErr) {
		t.Error("errors.Is should find base error through single wrapper")
	}

	if !IsConfigError(level1) {
		t.Error("IsConfigError should identify wrapped config error")
	}
}

func TestErrorChainPreservation(t *testing.T) {
	baseErr := errors.New("underlying")
	middleErr := NewRcloneError("rclone failed", baseErr)
	topErr := NewValidationError("validation failed", middleErr)

	if !errors.Is(topErr, baseErr) {
		t.Error("errors.Is should find base error through multiple wrappers")
	}

	if !errors.Is(topErr, middleErr) {
		t.Error("errors.Is should find middle error in chain")
	}

	var syncErr *SyncermanError
	if !errors.As(topErr, &syncErr) {
		t.Error("errors.As should find SyncermanError")
	}

	if syncErr.Type != TypeValidation {
		t.Errorf("expected TypeValidation at top level, got %v", syncErr.Type)
	}

	if !IsValidationError(topErr) {
		t.Error("IsValidationError should identify top-level error type")
	}
}

func TestErrorChainWithNilUnderlying(t *testing.T) {
	err := NewConfigError("message", nil)

	if errors.Is(err, errors.New("any")) {
		t.Error("errors.Is with nil underlying should not match random errors")
	}

	var syncErr *SyncermanError
	if !errors.As(err, &syncErr) {
		t.Error("errors.As should still work with nil underlying error")
	}
}

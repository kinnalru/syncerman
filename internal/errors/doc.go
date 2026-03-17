package errors

// Package errors provides custom error types and utilities for Syncerman.
//
// It defines domain-specific error types for configuration, rclone execution,
// and validation errors with proper error wrapping and formatting.
//
// The package provides:
//   - ErrorType enumeration for categorizing errors
//   - SyncermanError for structured error information
//   - Constructor functions (NewConfigError, NewRcloneError, NewValidationError)
//   - Type checking helpers (IsConfigError, IsRcloneError, IsValidationError)
//
// Error wrapping follows Go conventions, allowing use with errors.Is and errors.As.

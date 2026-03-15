package config

import (
	"fmt"
	"strings"

	"gitlab.com/kinnalru/syncerman/internal/errors"
)

// Validate performs comprehensive validation of the configuration structure.
// It ensures that all required fields are present and follow the expected format.
//
// Returns:
//   - error: A validation error with a descriptive message if validation fails,
//     nil if the configuration is valid.
//
// Validation checks performed:
//   - Configuration is not empty (Providers slice must contain at least one entry)
//   - Provider names are not empty
//   - Providers have at least one path defined
//   - Paths are not empty
//   - Each path has at least one destination
//   - Each destination is valid (see validateDestination for details)
//
// Example usage:
//
//	config := &Config{...}
//	if err := config.Validate(); err != nil {
//	    log.Fatalf("Invalid configuration: %v", err)
//	}
func (c *Config) Validate() error {
	// Ensure configuration has at least one provider defined
	if len(c.Providers) == 0 {
		return errors.NewValidationError("configuration is empty", nil)
	}

	for _, provider := range c.Providers {
		providerName := provider.Name
		paths := provider.Data

		// Provider names must not be empty
		if providerName == "" {
			return errors.NewValidationError("provider name cannot be empty", nil)
		}

		// Each provider must have at least one path defined
		if len(paths) == 0 {
			return errors.NewValidationError(fmt.Sprintf("provider %q has no paths defined", providerName), nil)
		}

		for _, pathData := range paths {
			path := pathData.Name
			destinations := pathData.Values

			// Source paths must not be empty
			if path == "" {
				return errors.NewValidationError(fmt.Sprintf("path cannot be empty for provider %q", providerName), nil)
			}

			// Each source path must have at least one destination
			if len(destinations) == 0 {
				return errors.NewValidationError(fmt.Sprintf("no destinations defined for provider %q path %q", providerName, path), nil)
			}

			// Validate each destination configuration
			for i, dest := range destinations {
				if err := validateDestination(providerName, path, dest, i); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateDestination validates a single destination configuration entry.
// It ensures the destination is properly formatted and contains all required fields.
//
// Parameters:
//   - provider: The name of the source provider (e.g., "gdrive1", "s3")
//   - path: The source path being synced
//   - dest: The Destination struct containing destination configuration
//   - index: The index of this destination in the destinations slice
//
// Returns:
//   - error: A validation error with a context-aware message including provider,
//     path, and index information if validation fails, nil if valid.
//
// Validation checks performed:
//   - Destination "to" field is not empty
//   - Destination format is valid (must be "provider:path" format or a local path
//     starting with "." or "/")
//   - All destination arguments are not empty
func validateDestination(provider string, path string, dest Destination, index int) error {
	if dest.To == "" {
		return errors.NewValidationError(fmt.Sprintf("destination 'to' field cannot be empty for provider %q path %q at index %d",
			provider, path, index), nil)
	}

	if !isValidDestinationFormat(dest.To) {
		return errors.NewValidationError(fmt.Sprintf("destination must be in format 'provider:path' or local path for provider %q path %q at index %d: %q",
			provider, path, index, dest.To), nil)
	}

	for _, arg := range dest.Args {
		if arg == "" {
			return errors.NewValidationError(fmt.Sprintf("destination argument cannot be empty for provider %q path %q at index %d",
				provider, path, index), nil)
		}
	}

	return nil
}

func isValidDestinationFormat(to string) bool {
	if strings.Contains(to, ":") {
		return true
	}
	if strings.HasPrefix(to, ".") || strings.HasPrefix(to, "/") {
		return true
	}
	return false
}

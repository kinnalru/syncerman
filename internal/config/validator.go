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
	if len(c.Providers) == 0 {
		return errors.NewValidationError("configuration is empty: at least one provider must be defined. "+
			"Example: providers:\n  gdrive:\n    \"/path\":\n      - to: \"remote:backup\"", nil)
	}

	for _, provider := range c.Providers {
		providerName := provider.Name
		paths := provider.Data

		if providerName == "" {
			return errors.NewValidationError("provider name cannot be empty. "+
				"Each provider must have a valid name (e.g., 'gdrive', 'dropbox', 's3'). "+
				"Provider names must correspond to rclone remotes defined in ~/.config/rclone/rclone.conf", nil)
		}

		if len(paths) == 0 {
			return errors.NewValidationError(fmt.Sprintf("provider %q has no paths defined. "+
				"Each provider must specify at least one source path to sync from. "+
				"Example:\n  %q:\n    \"/source/path\":\n      - to: \"destination:backup\"", providerName, providerName), nil)
		}

		for _, pathData := range paths {
			path := pathData.Name
			destinations := pathData.Values

			if path == "" {
				return errors.NewValidationError(fmt.Sprintf("source path cannot be empty for provider %q. "+
					"Each path key must specify a source location to sync from. "+
					"Example:\n  %q:\n    \"/documents\":\n      - to: \"backup:docs\"", providerName, providerName), nil)
			}

			if len(destinations) == 0 {
				return errors.NewValidationError(fmt.Sprintf("no destinations defined for provider %q path %q. "+
					"Each source path must have at least one destination to sync to. "+
					"Example:\n  %q:\n    %q:\n      - to: \"remote:backup\"",
					providerName, path, providerName, path), nil)
			}

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
		return errors.NewValidationError(fmt.Sprintf("destination 'to' field cannot be empty for provider %q path %q (destination #%d). "+
			"Each destination must specify where to sync to. "+
			"Valid formats:\n  - Remote provider: 'provider:path' (e.g., 'gdrive:backup/docs')\n  - Local path: '/absolute/path' or './relative/path'",
			provider, path, index), nil)
	}

	if !isValidDestinationFormat(dest.To) {
		return errors.NewValidationError(fmt.Sprintf("invalid destination format %q for provider %q path %q (destination #%d). "+
			"Destination must be one of:\n  - Remote provider with colon: 'provider:path' (e.g., 'gdrive:backup/docs', 'dropbox:archive')\n  - Absolute path starting with '/': '/absolute/path/to/backup'\n  - Relative path starting with '.' or '..': './backup' or '../backup'\n  "+
			"Provider names must correspond to rclone remotes defined in rclone configuration",
			dest.To, provider, path, index), nil)
	}

	for i, arg := range dest.Args {
		if arg == "" {
			return errors.NewValidationError(fmt.Sprintf("destination args element #%d cannot be empty for provider %q path %q (destination #%d). "+
				"All rclone arguments must be non-empty. If you don't need arguments, remove the 'args' field or provide a list of valid flags (e.g., args: ['--fast-list', '--max-age 30d'])",
				i, provider, path, index), nil)
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

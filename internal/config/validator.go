package config

import (
	"strings"

	"syncerman/internal/errors"
)

func (c *Config) Validate() error {
	if c.Providers == nil || len(c.Providers) == 0 {
		return errors.NewValidationError("configuration is empty", nil)
	}

	for providerName, paths := range c.Providers {
		if providerName == "" {
			return errors.NewValidationError("provider name cannot be empty", nil)
		}

		if len(paths) == 0 {
			return errors.NewValidationError("provider "+providerName+" has no paths defined", nil)
		}

		for path, destinations := range paths {
			if path == "" {
				return errors.NewValidationError("path cannot be empty for provider "+providerName, nil)
			}

			if len(destinations) == 0 {
				return errors.NewValidationError("no destinations defined for provider "+providerName+" path "+path, nil)
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

func validateDestination(provider string, path string, dest Destination, index int) error {
	if dest.To == "" {
		return errors.NewValidationError("destination 'to' field cannot be empty at provider "+provider+" path "+path+" index "+string(rune('0'+index)), nil)
	}

	if !strings.Contains(dest.To, ":") && !strings.HasPrefix(dest.To, ".") && !strings.HasPrefix(dest.To, "/") {
		return errors.NewValidationError("destination must be in format 'provider:path' or local path at provider "+provider+" path "+path+" index "+string(rune('0'+index)), nil)
	}

	for _, arg := range dest.Args {
		if arg == "" {
			return errors.NewValidationError("destination argument cannot be empty at provider "+provider+" path "+path+" index "+string(rune('0'+index)), nil)
		}
	}

	return nil
}

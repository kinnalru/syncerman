package config

// Package config provides configuration loading, discovery, and validation for Syncerman.
//
// This package handles the complete configuration lifecycle including file discovery, YAML parsing,
// validation of configuration structures, and retrieval of sync targets. Configuration files define
// sync targets between different storage providers using rclone backend interfaces.
//
// Configuration File Discovery
//
// Syncerman searches for configuration files using a priority-based discovery mechanism:
//
// 1. Explicit path: If a path is provided via command-line --config flag, that file is used
// 2. Default search: When no explicit path is given, the following search order is used:
//    - Current directory first, then parent directories up the filesystem tree
//    - In each directory, searches for files in this order: configuration.yml, config.yml, .syncerman.yml
//
// The discover process stops at the first valid configuration file found. If no file is located,
// a configuration error is returned describing the search locations.
//
// Configuration Loading
//
// Configuration is loaded from YAML files using the LoadConfig function, which reads the file
// contents and unmarshals them into internal data structures. The configuration format uses
// nested mappings: providers map to paths, and paths map to arrays of destinations.
//
// For loading from existing data, LoadConfigFromData can be used to parse YAML byte slices.
//
// Validation Rules
//
// After loading, configurations must pass validation before use. The validation process checks:
//
// - Overall structure: at least one provider must be defined
// - Provider names: must be non-empty strings
// - Provider paths: each provider must have at least one path defined
// - Path values: must be non-empty strings
// - Destination arrays: each path must have at least one destination
// - Destination "to" field: must be non-empty and in format "provider:path" or local path
//     (starting with "." or "/")
// - Argument arrays: all arguments must be non-empty strings
//
// Configuration validation errors include detailed context about which provider, path, or
// destination caused the validation failure.
//
// Error Handling
//
// The package uses typed error wrapping through the internal/errors package:
//
// - ConfigError: indicates failures during file reading or YAML parsing
// - ValidationError: indicates configuration structure violations
//
// All errors provide descriptive messages with context to help identify and resolve configuration
// issues. Validation is typically performed immediately after loading to detect configuration
// problems before sync operations begin.
//
// For detailed information about the configuration format and schema, including examples,
// see guides/OVERALL.md:46-111.

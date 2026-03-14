# Milestone 2: Document internal/config/*

## Goal
Enhance documentation for all config package files, including function-level documentation for all exported types and functions.

## Context

Current state analysis:
- internal/config/ has basic doc.go
- Functions need detailed documentation
- Config types need complete comments
- Validation logic requires documentation

Reference documents:
- `guides/OVERALL.md:46-111` - Configuration format and schema
- `guides/STYLE.md` - Go code style guidelines

## Tasks

### Task 2.1: Enhance config.go doc.go
Update package documentation to explain:
- Configuration loading mechanism
- File discovery logic
- Validation rules
- Error handling approach

### Task 2.2: Document Config struct
Add comprehensive documentation for Config struct:
- Purpose and usage
- Field descriptions
- Validation rules
- Example initialization

### Task 2.3: Document SyncTarget struct
Add comprehensive documentation for SyncTarget struct:
- Represents a single sync target
- Source and destination format
- Optional fields (args, resync)
- Usage examples

### Task 2.4: Document LoadConfig function
Add function documentation for LoadConfig:
- Purpose - loads configuration from file
- Parameters - file path
- Returns - Config struct and error
- Error cases - file not found, invalid YAML, validation errors
- Example usage

### Task 2.5: Document validateConfig function
Add function documentation for validateConfig:
- Purpose - validates configuration structure
- Validation rules applied
- Error messages returned
- Configuration requirements

### Task 2.6: Document all validation helper functions
Add documentation for:
- validateSyntax
- validateStructure
- validateProviders
- validateTargets

### Task 2.7: Add inline comments for complex logic
Document complex validation logic:
- Provider name validation
- Path format validation
- Remote format validation
- Optional field type checking

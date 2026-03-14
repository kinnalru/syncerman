# AGENTS.md

This document provides guidelines for autonomous coding agents working on this repository. Project name `Syncerman`.

> `Syncerman` is a console application (CLI) for syncronizing targets (sources and destination) based on `rclone` CLI.

## Current Project Guides

- `guides/OVERALL.md` - Comprehensive and detailed project description, features and corner cases. READ WHEN PLANING
- `guides/STYLE.md` - Go Code Style Guidelines
- `guides/PLANING.md` - Autonomous Coding Agent Workflow


## Technology Stack

- **Language**: Go 1.21+
- **Configuration Format**: YAML
- **Sync Execution**: Direct exec.Command (Unison binary)

## Build Commands

##### 1. Linux Go Compilation
```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/syncerman-linux-amd64
```

##### 2. Windows Go Compilation
```bash
# Build for Windows
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/syncerman-windows-amd64
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting and Formatting
```bash
# Format all Go files
go fmt ./...

# Run go vet for static analysis
go vet ./...

# Run goimports (install with: go install golang.org/x/tools/cmd/goimports@latest)
goimports -w .

# Run golangci-lint (install from https://golangci-lint.run/)
golangci-lint run

# Run staticcheck (optional)
staticcheck ./...
```

## Makefile Commands

A `Makefile` is provided to simplify common tasks:

- `make build`: Compiles the binary for Linux and Windows.
- `make test`: Runs all tests with coverage and verbose output.
- `make lint`: Runs `golangci-lint` and `go vet`.
- `make fmt`: Runs `go fmt` and `goimports`.
- `make clean`: Removes generated binaries.
- `make all`: Runs `make test` followed by `make build`.


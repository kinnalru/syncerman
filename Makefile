.PHONY: all build linux windows test lint fmt clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build parameters
BINARY_NAME=syncerman
LINUX_BINARY=$(BINARY_NAME)-linux-amd64
WINDOWS_BINARY=$(BINARY_NAME)-windows-amd64

# Version parameters
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")
GitCommit ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BuildTime ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")
GoVersion ?= $(shell go version 2>/dev/null | awk '{print $$3}' || echo "unknown")
LDFLAGS ?= -s -w -X gitlab.com/kinnalru/syncerman/internal/version.Version=$(VERSION) -X gitlab.com/kinnalru/syncerman/internal/version.GitCommit=$(GitCommit) -X gitlab.com/kinnalru/syncerman/internal/version.BuildTime=$(BuildTime) -X gitlab.com/kinnalru/syncerman/internal/version.GoVersion=$(GoVersion)

all: test build

build: linux windows

build: linux windows

linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(LINUX_BINARY)

windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(WINDOWS_BINARY)

test:
	$(GOTEST) -v -cover ./...

lint:
	@if command -v golangci-lint >/dev/null; then golangci-lint run; else echo "golangci-lint not installed, skipping..."; fi
	$(GOVET) ./...

fmt:
	$(GOFMT) ./...
	goimports -w .

clean:
	$(GOCLEAN)
	rm -f bin/$(LINUX_BINARY)
	rm -f bin/$(WINDOWS_BINARY)

.PHONY: all build test lint fmt clean

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

all: test build

build:
	# Build for Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "-s -w" -o bin/$(LINUX_BINARY)
	# Build for Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "-s -w" -o bin/$(WINDOWS_BINARY)

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

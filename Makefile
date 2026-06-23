# Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
# Date: 2026-06-22

BINARY_NAME=harbor
BUILD_DIR=build
MODULE=$(shell go list -m)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-X $(MODULE)/cmd.version=$(VERSION)

.PHONY: build test test-verbose test-cover test-cover-html clean install vet fmt lint tidy cross-build run help

## build: Compile the binary for the current platform
build:
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .

## test: Run all unit tests
test:
	go test ./... -count=1

## test-verbose: Run all unit tests with verbose output
test-verbose:
	go test ./... -v -count=1

## test-cover: Run tests with coverage report
test-cover:
	go test ./... -coverprofile=coverage.out
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out

## test-cover-html: Open coverage report in browser
test-cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "HTML report: coverage.html"

## clean: Remove build artifacts and coverage files
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## install: Build and install the binary to GOPATH/bin
install:
	go install -ldflags="$(LDFLAGS)" .

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format all Go files
fmt:
	go fmt ./...

## lint: Run fmt and vet together
lint: fmt vet

## tidy: Tidy and verify module dependencies
tidy:
	go mod tidy
	go mod verify

## cross-build: Build for all release platforms
cross-build: clean
	mkdir -p dist
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe .

## run: Build and run with optional ARGS (e.g. make run ARGS="notes list --json")
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

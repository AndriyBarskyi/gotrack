.PHONY: setup build run clean test test-coverage lint install help

.DEFAULT_GOAL := help

GO = go
BINARY_NAME = gotrack
BUILD_DIR = ./cmd/gotrack

help:
	@echo "GoTrack Makefile"
	@echo "================"
	@echo "Available commands:"
	@echo "  make setup         - Set up the project (install dependencies)"
	@echo "  make build         - Build the project"
	@echo "  make run           - Run the application"
	@echo "  make clean         - Clean up temporary files and caches"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make lint          - Run linting tools"
	@echo "  make install       - Install the binary to GOPATH/bin"

setup:
	$(GO) mod tidy
	$(GO) mod download

build:
	$(GO) build -o $(BINARY_NAME) $(BUILD_DIR)

run:
	$(GO) run $(BUILD_DIR)/main.go

test:
	$(GO) test ./...

test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

install:
	$(GO) install $(BUILD_DIR)

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	find . -name "*.test" -type f -delete
	find . -name "*.test.exe" -type f -delete
	find . -name "*.test.dSYM" -type d -delete
	find . -name "*.prof" -type f -delete
	find . -name "*.cov" -type f -delete

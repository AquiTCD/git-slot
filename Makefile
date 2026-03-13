BINARY_NAME := git-slot
PACKAGE := github.com/AquiTCD/git-slot
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u '+%Y-%m-%d')
LDFLAGS := -s -w \
	-X $(PACKAGE)/internal/cmd.version=$(VERSION) \
	-X $(PACKAGE)/internal/cmd.commit=$(COMMIT) \
	-X $(PACKAGE)/internal/cmd.date=$(DATE)

.DEFAULT_GOAL := help

## Build

.PHONY: build
build: ## Build binary to ./bin/
	go build -ldflags "$(LDFLAGS)" -o ./bin/$(BINARY_NAME) $(PACKAGE)/cmd/$(BINARY_NAME)

.PHONY: install
install: ## Install binary to $GOBIN or $GOPATH/bin
	go install -ldflags "$(LDFLAGS)" $(PACKAGE)/cmd/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $$(go env GOBIN)"
	@if [ -z "$$(go env GOBIN)" ]; then echo "Installed to $$(go env GOPATH)/bin"; fi

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf ./bin coverage.out

## Test

.PHONY: test
test: ## Run tests with race detector
	go test -race -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: test-coverage-html
test-coverage-html: test-coverage ## Open coverage report in browser
	go tool cover -html=coverage.out

## Lint

.PHONY: fmt
fmt: ## Format code
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

## All-in-one

.PHONY: check
check: fmt vet lint test ## Run fmt, vet, lint, and test

.PHONY: mod
mod: ## Tidy and verify modules
	go mod tidy
	go mod verify

## Help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

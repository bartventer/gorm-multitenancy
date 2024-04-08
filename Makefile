.SHELLFLAGS = -ecuo pipefail
SHELL = /bin/bash

# Variables
PKG_NAME := main
ENVFILE_PATH := ./.devcontainer/.env
COVERAGE_PROFILE := cover.out
BINARY := $(PKG_NAME)
DEPS_SCRIPT := ./internal/testing/start_local_deps.sh

# Commands 
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOVET := go vet
GOTEST := go test
GOCOVER := go tool cover

# Flags
GOFLAGS := -v
GOFMTFLAGS := -s
GOLINTFLAGS := run --verbose
GOVETFLAGS := -all
GOTESTFLAGS := -v
GOCOVERFLAGS := -html

# Include environment variables
include $(ENVFILE_PATH)
export $(shell sed 's/=.*//' $(ENVFILE_PATH))

.PHONY: help
help: ## Display this help message.
	@echo "Usage: make [TARGET]"
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m    %-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Variables:"
	@awk 'BEGIN {FS = "##"} /^[a-zA-Z_-]+\s*\?=\s*.*?## / {split($$1, a, "\\s*\\?=\\s*"); printf "\033[33m   %-30s\033[0m %s\n", a[1], $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Variable Values:"
	@awk 'BEGIN {FS = "[ ?=]"} /^[a-zA-Z_-]+[ \t]*[?=]/ {print $$1}' $(MAKEFILE_LIST) | \
	while read -r var; do \
		printf "\033[35m    %-30s\033[0m %s\n" "$$var" "$$(make -s -f $(firstword $(MAKEFILE_LIST)) print-$$var)"; \
	done

.PHONY: print-%
print-%: ## Helper target to print a variable. Usage: make print-VARIABLE
	@printf '%s' "$($*)" | sed 's/^[[:space:]]*//'

.PHONY: all
all: build test ## Run all targets

.PHONY: fmt
fmt: ## Run gofmt on all files
	$(GOFMT) $(GOFMTFLAGS) -w .

.PHONY: lint
lint: ## Run golint on all files
	$(GOLINT) $(GOLINTFLAGS) ./...

.PHONY: vet
vet: ## Run go vet on all files
	$(GOVET) $(GOVETFLAGS) ./...

.PHONY: build
build: vet
	$(GO) build $(GOFLAGS) -o $(BINARY)

.PHONY: test
test: vet ## Run tests
	$(DEPS_SCRIPT)
	$(GOTEST) $(GOTESTFLAGS) -coverprofile=$(COVERAGE_PROFILE) `go list ./... | grep -v ./internal`

.PHONY: cover
cover: test ## Run tests with coverage
	$(GOCOVER) $(GOCOVERFLAGS) $(COVERAGE_PROFILE)
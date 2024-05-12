.SHELLFLAGS = -ecuo pipefail
SHELL = /bin/bash

# Variables
PKG_NAME := main
ENVFILE_PATH := ./.devcontainer/dev.env
COVERAGE_PROFILE := cover.out
BINARY := $(PKG_NAME)
SCRIPTS_DIR := ./scripts
DEPS_SCRIPT := $(SCRIPTS_DIR)/start_local_deps.sh
DEPS_HEALTH_SCRIPT := $(SCRIPTS_DIR)/check_services.sh
BENCH_SCRIPT := $(SCRIPTS_DIR)/run_benchmarks.sh
SKIP_DEPS ?= false
EXAMPLE_BUILDTAG := gormmultitenancy_example

# Commands 
GO := go
GOLANGCILINT := golangci-lint
GOTEST := $(GO) test
GOCOVER := $(GO) tool cover

# Flags
GOFLAGS := -v
GOLANGCILINTFLAGS := run --verbose
ifeq ($(CI),)
	GOLANGCILINTFLAGS += --fix --color always
endif
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

.PHONY: lint
lint: ## Run golint on all files
	$(GOLANGCILINT) $(GOLANGCILINTFLAGS) ./...

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o $(BINARY)

.PHONY: deps
deps: ## Start local dependencies
ifneq ($(SKIP_DEPS), true)
	$(DEPS_SCRIPT)
	$(DEPS_HEALTH_SCRIPT) "pgadmin"
endif

.PHONY: test
test: deps ## Run tests
	$(GOTEST) $(GOTESTFLAGS) -coverprofile=$(COVERAGE_PROFILE) ./...

BENCH_OUTDIR := ./tmp/bench
.PHONY: bench
bench: deps ## Run benchmarks
	$(BENCH_SCRIPT) \
		-package ./drivers/postgres/schema \
		-benchfuncs "BenchmarkScopingQueries" \
		-outputdir $(BENCH_OUTDIR)

.PHONY: cover
cover: test ## Run tests with coverage
	$(GOCOVER) $(GOCOVERFLAGS) $(COVERAGE_PROFILE)

.PHONY: echo_example
echo_example: deps## Run the echo example
	$(GO) run -C ./examples/echo -tags $(EXAMPLE_BUILDTAG) .

.PHONY: nethttp_example
nethttp_example: deps ## Run the nethttp example
	$(GO) run -C ./examples/nethttp -tags $(EXAMPLE_BUILDTAG) .

.PHONY: update
update: ## Update dependencies
	$(SCRIPTS_DIR)/update_deps.sh -tags $(EXAMPLE_BUILDTAG)
# Copyright © 2023 Bart Venter <bartventer@outlook.com>

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.SHELLFLAGS = -ecuo pipefail
SHELL = /bin/bash

# Arguments
SKIP_DEPS ?= false
COVER ?= false

# Variables
PKG_NAME := main
BINARY := $(PKG_NAME)
ENVFILE_PATH := ./.devcontainer/.env.dev

# Scripts
SCRIPTS_DIR := ./scripts
DEPS_SCRIPT := $(SCRIPTS_DIR)/start_local_deps.sh
DEPS_HEALTH_SCRIPT := $(SCRIPTS_DIR)/check_services.sh
BENCH_SCRIPT := $(SCRIPTS_DIR)/benchmark.sh

# Commands 
GO := go
GOLANGCILINT := golangci-lint
GOCOVER := $(GO) tool cover

# Flags
GOFLAGS := -v
GOLANGCILINTFLAGS := run --verbose
ifeq ($(CI),)
	GOLANGCILINTFLAGS += --fix --color always
endif

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
	$(SCRIPTS_DIR)/test.sh

.PHONY: benchmark
benchmark: deps ## Run benchmarks
	$(BENCH_SCRIPT) \
		-package ./drivers/postgres/schema \
		-benchfunc "BenchmarkScopingQueries" \
		-outputdir ./drivers/postgres/docs \
		-template $(SCRIPTS_DIR)/benchmark_template.md

.PHONY: update_readme
update_readme: ## Update the postgres driver README
	$(SCRIPTS_DIR)/update_readme.sh \
		--dirpath ./drivers/postgres/docs \
		--outfile ./drivers/postgres/README.md

.PHONY: coverbrowser
coverbrowser: ## View coverage in browser
	$(GOCOVER) -html=$(COVERAGE_PROFILE)

define run_example
	$(GO) run -C ./examples/$(1) .
endef

.PHONY: echo_example
examples/echo: deps ## Run the echo example
	$(call run_example,echo)

.PHONY: nethttp_example
examples/nethttp: deps ## Run the nethttp example
	$(call run_example,nethttp)

.PHONY: update_deps
update_deps: ## Update dependencies
	$(SCRIPTS_DIR)/update_deps.sh
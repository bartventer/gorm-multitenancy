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

SHELL = /bin/bash
.SHELLFLAGS = -ecuo pipefail

SCRIPTS_DIR := ./scripts

.PHONY: lint
lint: ## Lint the project
	$(SCRIPTS_DIR)/lint.sh

.PHONY: fmt
fmt: ## Format the project
	$(SCRIPTS_DIR)/fmt.sh

.PHONY: build
build: ## Build the project
	$(SCRIPTS_DIR)/build.sh

.PHONY: test
test: ## Run tests
	$(SCRIPTS_DIR)/test.sh

.PHONY: benchmark
benchmark: ## Run benchmarks
	$(SCRIPTS_DIR)/benchmark.sh

.PHONY: coverbrowser
coverbrowser: ## View coverage in browser
	go tool cover -html=coverage.out

.PHONY: update_deps
update_deps: ## Update dependencies
	$(SCRIPTS_DIR)/update_deps.sh

.PHONY: release
release: ## Create a new release
	$(SCRIPTS_DIR)/release.sh
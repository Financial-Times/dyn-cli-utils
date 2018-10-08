SHELL := /bin/bash

TEAL = $(shell printf '%b' "\033[0;36m")
GREEN = $(shell printf '%b' "\033[0;32m")
RED = $(shell printf '%b' "\033[0;31m")
NO_COLOUR = $(shell printf '%b' "\033[m")

PACKAGES = $(shell go list ./... | grep -v /vendor/)
DONE = printf '%b\n' ">> $(GREEN)$@ done ✓"

.PHONY: docs
docs: ## Automatically generate markdown documentation using Cobra
	@printf '%b\n' ">> $(TEAL)generating docs"
	go run cmd/generatedocs/generatedocs.go
	@$(DONE)

.PHONY: style
style: ## Check the formatting of the Go source code.
	@printf '%b\n' ">> $(TEAL)checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'
	@$(DONE)

.PHONY: format
format: ## Format the Go source code.
	@printf '%b\n' ">> $(TEAL)formatting code"
	go fmt $(PACKAGES)
	@$(DONE)

.PHONY: vet
vet: ## Examine the Go source code.
	@printf '%b\n' ">> $(TEAL)vetting code"
	go vet $(PACKAGES)
	@$(DONE)

.PHONY: help
help: ## Show this help message.
	@printf '%b\n' "usage: make [target] ..."
	@printf '%b\n' ""
	@printf '%b\n' "targets:"
	@grep -Eh '^.+:\ ##\ .+' ${MAKEFILE_LIST} | column -t -s ':#'

SHELL := /bin/bash

TEAL = $(shell printf '%b' "\033[0;36m")
GREEN = $(shell printf '%b' "\033[0;32m")
RED = $(shell printf '%b' "\033[0;31m")
NO_COLOUR = $(shell printf '%b' "\033[m")

PACKAGES = $(shell go list ./... | grep -v /vendor/)
DONE = printf '%b\n' ">> $(GREEN)$@ done ✓"

docs: $(wildcard cmd/**/*.go) ## Automatically generate markdown documentation using Cobra
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

.PHONY: security
security: _security-login ## Scan dependencies for security vulnerabilities.
	# TODO: enable once snyk support go modules https://github.com/snyk/snyk/issues/354
	# @printf '%b\n' ">> $(TEAL)scanning dependencies for vulnerabilities"
	# npx snyk test --org=reliability-engineering
	# @$(DONE)

_security-login:

_security-login-web: ## Login to snyk if not on CI.
	# TODO: enable once snyk support go modules https://github.com/snyk/snyk/issues/354
	# @printf '%b\n' ">> $(TEAL)Not on CI, logging into Snyk"
	# npx snyk auth

ifeq ($(CI),)
_security-login: _security-login-web
endif

.PHONY: security-monitor
security-monitor: ## Update latest monitored dependencies in snyk. Needs to be run in an environment with the snyk CLI tool.
	# TODO: enable once snyk support go modules https://github.com/snyk/snyk/issues/354
	# @printf '%b\n' ">> $(TEAL)updating snyk dependencies"
	# npx snyk monitor --org=reliability-engineering
	# @$(DONE)

.PHONY: help
help: ## Show this help message.
	@printf '%b\n' "usage: make [target] ..."
	@printf '%b\n' ""
	@printf '%b\n' "targets:"
	@# replace the first : with £ to avoid splitting columns on URLs
	@grep -Eh '^[^_].+?:\ ##\ .+' ${MAKEFILE_LIST} | cut -d ' ' -f '1 3-' | sed 's/^(.+?):/$1/' | sed 's/:/£/' | column -t -c 2 -s '£'

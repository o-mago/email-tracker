PROJECT ?= go-api-template

BUILD_TIME=$(shell date '+%Y-%m-%d__%I:%M:%S%p')
GIT_COMMIT=$(shell git rev-parse HEAD)
TAG=$(shell git rev-parse --abbrev-ref HEAD)

DOCKER_COMPOSE_FILE=build/docker-compose.yml

GOLANGCI_LINT_PATH=$$(go env GOPATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION=1.58.1

.PHONY: *

lint:
	@echo "==> Installing golangci-lint"
ifeq (,$(findstring $(GOLANGCI_LINT_VERSION),$(shell which $(GOLANGCI_LINT_PATH) && eval $(GOLANGCI_LINT_PATH) version)))
	@echo "installing golangci-lint v$(GOLANGCI_LINT_VERSION)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v$(GOLANGCI_LINT_VERSION)
else
	@echo "already installed: $(shell eval $(GOLANGCI_LINT_PATH) version)"
endif
	@echo "==> Running golangci-lint"
	@$(GOLANGCI_LINT_PATH) run -c ./.golangci.yml --fix

start:
	docker-compose -f $(DOCKER_COMPOSE_FILE) -p $(PROJECT) down --remove-orphans
	docker-compose -f $(DOCKER_COMPOSE_FILE) -p $(PROJECT) up --remove-orphans

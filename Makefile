COMPONENTS = agent server staticlint
E2E_TEST = test/devopstest
API_DOCS = docs/api

AGENT_VERSION ?= 0.19.0
SERVER_VERSION ?= 0.19.0

BUILD_DATE ?= $(shell date +%F\ %H:%M:%S)
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD)

DEFAULT_GOAL := help

help: ## Display this help screen
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

install-tools: $(E2E_TEST) ## Install additional linters and test tools
	pre-commit install
.PHONY: install-tools

$(E2E_TEST):
	@echo Installing $@
	curl -sSfL \
		https://github.com/Yandex-Practicum/go-autotests/releases/download/v0.7.12/devopstest-darwin-amd64 \
		-o $@
	@chmod +x $(E2E_TEST)

build: $(COMPONENTS) ## Build whole project
.PHONY: build

agent: ## Build agent
	go build \
		-ldflags "\
			-X 'main.buildVersion=$(AGENT_VERSION)' \
			-X 'main.buildDate=$(BUILD_DATE)' \
			-X 'main.buildCommit=$(BUILD_COMMIT)' \
		" \
		-o cmd/$@/$@ \
		cmd/$@/*.go
.PHONY: agent

server: ## Build metrics server
	rm -rf $(API_DOCS)
	swag init -g ./internal/handlers/router.go --output $(API_DOCS)

	go build \
		-ldflags "\
			-X 'main.buildVersion=$(SERVER_VERSION)' \
			-X 'main.buildDate=$(BUILD_DATE)' \
			-X 'main.buildCommit=$(BUILD_COMMIT)' \
		" \
		-o cmd/$@/$@ \
		cmd/$@/*.go
.PHONY: server

staticlint: ## Build static lint utility
	go build $(CCFLAGS) -o cmd/$@/$@ cmd/$@/*.go
.PHONY: staticlint

clean: ## Remove build artifacts and downloaded test tools
	rm -rf cmd/agent/agent cmd/server/server cmd/staticlint/staticlint $(E2E_TEST)
.PHONY: clean

lint: ## Run linters on the source code
	golangci-lint run
	./cmd/staticlint/staticlint ./cmd/... ./internal/... ./pkg/...
.PHONY: lint

unit-tests: ## Run unit tests
	@go test -v -race ./... -coverprofile=coverage.out.tmp -covermode atomic
	@cat coverage.out.tmp | grep -v "_mock.go" > coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out
.PHONY: unit-tests

e2e-tests: $(E2E_TEST) build ### Run e2e tests
	./scripts/run-e2e-test
.PHONY: e2e-tests

godoc: ### Show public packages documentation using godoc
	@echo "Project documentation is available at:"
	@echo "http://127.0.0.1:3000/pkg/github.com/alkurbatov/metrics-collector/pkg/\n"
	@godoc -http=:3000 -index -play
.PHONY: godoc

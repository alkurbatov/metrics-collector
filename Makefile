COMPONENTS = agent server
E2E_TEST = test/devopstest
CCFLAGS =

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

$(COMPONENTS):
	go build $(CCFLAGS) -o cmd/$@/$@ cmd/$@/*.go

clean: ## Remove build artifacts and downloaded test tools
	rm -f cmd/agent/agent cmd/server/server $(E2E_TEST)
.PHONY: clean

lint: ## Run linters on the source code
	golangci-lint run
.PHONY: lint

unit-tests: ## Run unit tests
	@go test -v -race ./... -coverprofile=coverage.out.tmp -covermode atomic
	@cat coverage.out.tmp | grep -v "_mock.go" > coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out
.PHONY: unit-tests

e2e-tests: $(E2E_TEST) ### Run e2e tests
	./scripts/run-e2e-test
.PHONY: e2e-tests

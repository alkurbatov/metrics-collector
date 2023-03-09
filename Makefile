COMPONENTS = agent server
E2E_TEST = test/devopstest
API_DOCS = docs/api
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

build: api-docs $(COMPONENTS) ## Build whole project
.PHONY: build

$(COMPONENTS):
	go build $(CCFLAGS) -o cmd/$@/$@ cmd/$@/*.go

api-docs:
	rm -rf $(API_DOCS)
	swag init -g ./internal/handlers/router.go --output $(API_DOCS)
.PHONY: api-docs

clean: ## Remove build artifacts and downloaded test tools
	rm -f cmd/agent/agent cmd/server/server $(E2E_TEST) $(API_DOCS)
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

e2e-tests: $(E2E_TEST) build ### Run e2e tests
	./scripts/run-e2e-test
.PHONY: e2e-tests

godoc: ### Show public packages documentation using godoc
	@echo "Project documentation is available at:"
	@echo "http://127.0.0.1:3000/pkg/github.com/alkurbatov/metrics-collector/pkg/metrics/\n"
	@godoc -http=:3000 -index -play
.PHONY: godoc

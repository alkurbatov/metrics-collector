COMPONENTS = agent server
E2E_TEST = tests/devopstest
CCFLAGS =
PG_URL = 'postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable'

DEFAULT_GOAL := help

help: ## Display this help screen
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

install-tools: $(E2E_TEST) ## Install additional test tools
.PHONY: install-tools

$(E2E_TEST):
	@echo Installing $@
	curl -sSfL \
		https://github.com/Yandex-Practicum/go-autotests/releases/download/v0.7.8/devopstest-darwin-amd64 \
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
	@go test -v ./... -coverprofile=coverage.out.tmp -covermode count
	@cat coverage.out.tmp | grep -v "_mock.go" > coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out
.PHONY: unit-tests

e2e-tests: $(E2E_TEST) ### Run e2e tests
	@rm -f /tmp/test_store*.json
	@$(E2E_TEST) -test.v -test.run=^TestIteration1$$ \
		-agent-binary-path=cmd/agent/agent
	@$(E2E_TEST) -test.v -test.run=^TestIteration2[b]*$$ \
		-source-path=. \
		-binary-path=cmd/server/server
	@$(E2E_TEST) -test.v -test.run=^TestIteration3[b]*$$ \
		-source-path=. \
		-binary-path=cmd/server/server
	@$(E2E_TEST) -test.v -test.run=^TestIteration4$$ \
		-source-path=. \
		-binary-path=cmd/server/server \
		-agent-binary-path=cmd/agent/agent
	@ADDRESS="localhost:5000" \
	$(E2E_TEST) -test.v -test.run=^TestIteration5$$ \
		-source-path=. \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=5000
	@ADDRESS="localhost:6000" \
	TEMP_FILE="/tmp/test_store6.json" \
	$(E2E_TEST) -test.v -test.run=^TestIteration6$$ \
		-source-path=. \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=6000 \
		-database-dsn=$(PG_URL) \
		-file-storage-path="/tmp/test_store6.json"
	@ADDRESS="localhost:7000" \
	TEMP_FILE="/tmp/test_store7.json" \
	$(E2E_TEST) -test.v -test.run=^TestIteration7$$ \
		-source-path=. \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=7000 \
		-database-dsn=$(PG_URL) \
		-file-storage-path="/tmp/test_store7.json"
	@ADDRESS="localhost:8000" \
	TEMP_FILE="/tmp/test_store8.json"
	$(E2E_TEST) -test.v -test.run=^TestIteration8$$ \
		-source-path=. \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=8000 \
		-database-dsn=$(PG_URL) \
		-file-storage-path=/tmp/test_store8.json
.PHONY: e2e-tests

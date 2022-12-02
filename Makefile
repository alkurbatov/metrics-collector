COMPONENTS = agent server
CCFLAGS =

DEFAULT_GOAL := help

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

build: $(COMPONENTS) ## Build whole project
.PHONY: build

%:
	go build $(CCFLAGS) -o cmd/$@/$@ cmd/$@/*.go

clean: ## Cleanup build artifacts
	rm -f cmd/agent/agent cmd/server/server
.PHONY: clean

tests: ## Run unit tests
	@go test -v ./... -coverprofile=coverage.out.tmp -covermode count
	@cat coverage.out.tmp | grep -v "_mock.go" > coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out
.PHONY: tests

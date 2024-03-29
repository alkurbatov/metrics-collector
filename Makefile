COMPONENTS = agent server staticlint
E2E_TEST = test/devopstest
API_DOCS = docs/api
KEY_PATH = build/keys

PROTO_SRC = api/proto
PROTO_FILES = health metrics
PROTO_DST = pkg/grpcapi

AGENT_VERSION ?= 0.24.0
SERVER_VERSION ?= 0.24.0

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
		https://github.com/Yandex-Practicum/go-autotests/releases/download/v0.9.6/devopstest-darwin-amd64 \
		-o $@
	@chmod +x $(E2E_TEST)

build: proto $(COMPONENTS) ## Build whole project
.PHONY: build

proto: $(PROTO_FILES) ## Generate gRPC protobuf bindings
.PHONY: proto

$(PROTO_FILES): %: $(PROTO_DST)/%

$(PROTO_DST)/%:
	protoc \
		--proto_path=$(PROTO_SRC) \
		--go_out=$(PROTO_DST) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DST) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_SRC)/$(notdir $@).proto

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
	swag init -g ./internal/httpbackend/router.go --output $(API_DOCS)

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
	go build -o cmd/$@/$@ cmd/$@/*.go
.PHONY: staticlint

clean: ## Remove build artifacts and downloaded test tools
	rm -rf cmd/agent/agent cmd/server/server cmd/staticlint/staticlint $(E2E_TEST)
.PHONY: clean

lint: ## Run linters on the source code
	golangci-lint run
	go list -buildvcs=false ./... | grep -F -v -e docs | xargs ./cmd/staticlint/staticlint
.PHONY: lint

unit-tests: ## Run unit tests
	@go test -v -race ./... -coverprofile=coverage.out.tmp -covermode atomic
	@cat coverage.out.tmp | grep -v -E "(_mock|.pb).go" > coverage.out
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

keys: ## Generate private and public RSA key pair to encrypt agent -> server communications
	mkdir -p $(KEY_PATH)
	openssl genrsa -out $(KEY_PATH)/private.pem 4096
	openssl rsa -in $(KEY_PATH)/private.pem -outform PEM -pubout -out $(KEY_PATH)/public.pem
.PHONY: keys

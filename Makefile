COMPONENTS = agent server
CCFLAGS =

all: $(COMPONENTS)
.PHONY: all

%:
	go build $(CCFLAGS) -o cmd/$@/$@ cmd/$@/*.go

clean:
	rm -f cmd/agent/agent cmd/server/server
.PHONY: clean

tests:
	@go test -v ./... -coverprofile=coverage.out.tmp -covermode count
	@cat coverage.out.tmp | grep -v "_mock.go" > coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out
.PHONY: tests

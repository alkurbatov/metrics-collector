COMPONENTS = agent server
CCFLAGS =

all: $(COMPONENTS)

%:
	go build $(CCFLAGS) -o cmd/$@/$@ cmd/$@/*.go

clean:
	rm -f cmd/agent/agent cmd/server/server

tests:
	@go test -v ./... -coverprofile=coverage.out.tmp -covermode count
	@cat coverage.out.tmp | grep -v "_mock.go" > coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out

.PHONY: all clean tests

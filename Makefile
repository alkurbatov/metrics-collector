COMPONENTS = agent server
CCFLAGS =

all: $(COMPONENTS)

%:
	go build $(CCFLAGS) -o cmd/$@/$@ cmd/$@/*.go

clean:
	rm -f cmd/agent/agent cmd/server/server

.PHONY: all clean

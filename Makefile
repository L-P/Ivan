EXEC=./$(shell basename "$(shell pwd)")
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "unknown")
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'

all: $(EXEC)

$(EXEC):
	go build $(BUILDFLAGS)

.PHONY: $(EXEC) vendor upgrade lint test coverage

coverage:
	go test -tags docker,api -covermode=count -coverprofile=coverage.cov --timeout=30s ./...
	go tool cover -html=coverage.cov -o coverage.html
	rm coverage.cov
	sensible-browser coverage.html

test:
	go test ./...

vendor:
	go get -v
	go mod vendor
	go mod tidy

upgrade:
	go get -u -v
	go mod vendor
	go mod tidy

lint: $(GOLANGCI)
	golangci-lint run
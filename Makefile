.PHONY: all fmt vet test

all: vet test

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test:
	go test -race -coverprofile=/tmp/coverage.out ./... && go tool cover -func=/tmp/coverage.out

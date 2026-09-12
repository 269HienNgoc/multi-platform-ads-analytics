BINARY_NAME := ads-api
GO := go

.PHONY: build check clean fmt lint run test tidy vet

build:
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/api

check: fmt vet test build

clean:
	rm -rf bin coverage.out coverage.html

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run ./...

run:
	$(GO) run ./cmd/api -config configs/config.yaml

test:
	$(GO) test -race -shuffle=on -coverprofile=coverage.out ./...

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet ./...

BINARY_NAME := ads-api
GO := go

.PHONY: build check clean fmt lint migrate-down migrate-up migrate-version run test tidy vet

build:
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/api

check: fmt vet test build

clean:
	rm -rf bin coverage.out coverage.html

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run ./...

migrate-up:
	$(GO) run ./cmd/migrate -config configs/config.yaml -migrations migrations -action up

migrate-down:
	$(GO) run ./cmd/migrate -config configs/config.yaml -migrations migrations -action down -steps 1

migrate-version:
	$(GO) run ./cmd/migrate -config configs/config.yaml -migrations migrations -action version

run:
	$(GO) run ./cmd/api -config configs/config.yaml

test:
	$(GO) test -race -shuffle=on -coverprofile=coverage.out ./...

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet ./...

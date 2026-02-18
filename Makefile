SHELL := /bin/bash

.DEFAULT_GOAL := build

.PHONY: build run test test-cover fmt fmt-check lint vet staticcheck ci tidy

BIN_DIR := $(CURDIR)/bin
BIN := $(BIN_DIR)/sdo
CMD := ./cmd/sdo
GO ?= go

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo "")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/ra/scrape_do_cli/internal/cmd.version=$(VERSION) -X github.com/ra/scrape_do_cli/internal/cmd.commit=$(COMMIT) -X github.com/ra/scrape_do_cli/internal/cmd.date=$(DATE)

build:
	@mkdir -p $(BIN_DIR)
	@$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD)

run: build
	@$(BIN) $(ARGS)

test:
	@$(GO) test ./...

test-cover:
	@$(GO) test -covermode=atomic -coverprofile=coverage.out ./...
	@$(GO) tool cover -func=coverage.out | tail -n 1

fmt:
	@gofmt -w .

fmt-check:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then \
		echo "gofmt required for:"; \
		echo "$$out"; \
		exit 1; \
	fi

vet:
	@$(GO) vet ./...

staticcheck:
	@PATH="$$($(GO) env GOPATH)/bin:$$PATH"; \
	if ! command -v staticcheck >/dev/null 2>&1; then \
		echo "installing staticcheck..."; \
		$(GO) install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi; \
	staticcheck ./...

lint: fmt-check vet staticcheck

ci: lint test-cover build

tidy:
	@$(GO) mod tidy

SHELL := /bin/bash

.DEFAULT_GOAL := build

.PHONY: build run test fmt lint tidy

BIN_DIR := $(CURDIR)/bin
BIN := $(BIN_DIR)/sdo
CMD := ./cmd/sdo

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo "")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/ra/scrape_do_cli/internal/cmd.version=$(VERSION) -X github.com/ra/scrape_do_cli/internal/cmd.commit=$(COMMIT) -X github.com/ra/scrape_do_cli/internal/cmd.date=$(DATE)

build:
	@mkdir -p $(BIN_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD)

run: build
	@$(BIN) $(ARGS)

test:
	@go test ./...

fmt:
	@gofmt -w .

lint:
	@echo "lint placeholder"

tidy:
	@go mod tidy

APP_NAME := money-tracker
BUILD_DIR := bin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | awk '{print $$3}')

LDFLAGS := -X icekalt.dev/money-tracker/internal/buildinfo.Version=$(VERSION) \
           -X icekalt.dev/money-tracker/internal/buildinfo.Commit=$(COMMIT) \
           -X icekalt.dev/money-tracker/internal/buildinfo.BuildDate=$(BUILD_DATE) \
           -X icekalt.dev/money-tracker/internal/buildinfo.GoVersion=$(GO_VERSION)

.PHONY: build run test lint clean generate migrate

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/money-tracker

run: build
	./$(BUILD_DIR)/$(APP_NAME) serve

test:
	go test ./... -count=1

test-integration:
	go test ./tests/integration/... -count=1 -tags=integration

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

generate:
	go generate ./...

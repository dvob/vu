.PHONY: setup-lint lint test build

GOLANGCI_VERSION ?= 1.17.1
VERSION          ?= $(shell git describe --tags --always --dirty)
COMMIT           ?= $(shell git rev-parse HEAD)
LDFLAGS          ?= -X main.version=$(VERSION) -X main.commit=$(COMMIT) -w -s

setup-lint:
	curl -sSfL https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_VERSION)/golangci-lint-$(GOLANGCI_VERSION)-linux-amd64.tar.gz | \
		tar -xzv --wildcards --strip-components=1 -C $(shell go env GOPATH)/bin '*/golangci-lint'

lint: $(GOLANGCI_LINT)
	golangci-lint run

test:
	go test

build:
	go build -ldflags "$(LDFLAGS)"

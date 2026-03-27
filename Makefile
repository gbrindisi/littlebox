.PHONY: build test install setup

PREFIX ?= /usr/local

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT)

build:
	go build -ldflags '$(LDFLAGS)' -o bin/littlebox ./cmd/littlebox

install: build
	install -d $(PREFIX)/bin
	install -m 755 bin/littlebox $(PREFIX)/bin/littlebox

test:
	go test ./...

setup:
	go install github.com/securego/gosec/v2/cmd/gosec@latest

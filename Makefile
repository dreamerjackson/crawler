
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

VERSION := v1.0.0

CHANNEL := $(shell git rev-parse --abbrev-ref HEAD)
CHANNEL_BUILD = $(CHANNEL)-$(shell git rev-parse --short=7 HEAD)
project=github.com/dreamerjackson/crawler

LDFLAGS = -X "main.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "main.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "main.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
LDFLAGS += -X "main.Version=${VERSION}"

ifeq ($(gorace), 1)
	BUILD_FLAGS=-race
endif

build:
	go build -ldflags '$(LDFLAGS)' $(BUILD_FLAGS) main.go

lint:
	golangci-lint run ./...

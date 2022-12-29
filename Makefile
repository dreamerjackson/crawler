
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

VERSION := v1.0.0

CHANNEL := $(shell git rev-parse --abbrev-ref HEAD)
CHANNEL_BUILD = $(CHANNEL)-$(shell git rev-parse --short=7 HEAD)
project=github.com/dreamerjackson/crawler

LDFLAGS = -X "github.com/dreamerjackson/crawler/version.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/dreamerjackson/crawler/version.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "github.com/dreamerjackson/crawler/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
LDFLAGS += -X "github.com/dreamerjackson/crawler/version.Version=${VERSION}"

ifeq ($(gorace), 1)
	BUILD_FLAGS=-race
endif

build:
	go build -ldflags '$(LDFLAGS)' $(BUILD_FLAGS) main.go

debug:
	go build -gcflags=all="-N -l" -ldflags '$(LDFLAGS)' $(BUILD_FLAGS) main.go


lint:
	golangci-lint run ./...

imports:
	goimports -w .

cover:
	go test ./... -v -short -coverprofile .coverage.txt
	go tool cover -func .coverage.txt

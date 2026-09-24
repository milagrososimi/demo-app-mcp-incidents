VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)

.PHONY: all build vet test

all: build vet test

build:
	go build -ldflags "$(LDFLAGS)" ./...

vet:
	go vet ./...

test:
	go test ./...

.PHONY: build test lint clean install

CLI := ./cmd/goblocks

build:
	go build -o bin/goblocks $(CLI)

install:
	go install $(CLI)

test:
	go test ./... -race -count=1

test-integration:
	GOBLOCKS_PATH=$(GOBLOCKS_PATH) go test ./internal/scaffold/... -race -count=1

lint:
	go vet ./...
	go fmt ./...

clean:
	rm -rf bin/

.DEFAULT_GOAL := build

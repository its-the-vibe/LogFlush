.PHONY: build run test lint

build:
	go build -o bin/logflush ./cmd/logflush

run:
	go run ./cmd/logflush -config=config.yaml

test:
	go test ./...

lint:
	go vet ./...

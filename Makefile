.PHONY: build clean test run docker-build docker-run help

BINARY_NAME=guardian
VERSION?=0.1.0
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION)"

help:
	@echo "Guardian - Cloud-Native Process Daemon"
	@echo ""
	@echo "Available targets:"
	@echo "  make build        - Build the binary"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make run          - Run locally with example config"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run in Docker container"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Built $(BINARY_NAME)"

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

run: build
	./$(BINARY_NAME) -config guardian.yaml

docker-build:
	docker build -t guardian:$(VERSION) .
	@echo "Built Docker image guardian:$(VERSION)"

docker-run: docker-build
	docker run --rm -it \
		-p 9090:9090 \
		guardian:$(VERSION)

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

install-deps:
	go mod download
	go mod tidy

.DEFAULT_GOAL := help

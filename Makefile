.PHONY: build run clean test docker docker-run fmt lint install help cross version deploy

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)

BINARY := bin/snip
MAIN   := ./cmd/server

.DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(MAIN)

run: build ## Build and run
	$(BINARY)

clean: ## Clean build artifacts
	rm -rf bin/ dist/

test: ## Run tests
	go test ./... -v -count=1

cover: ## Run tests with coverage
	go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

bench: ## Run benchmarks
	go test ./... -bench=. -benchmem

fmt: ## Format code
	go fmt ./...

lint: ## Lint code
	go vet ./...

docker: ## Build Docker image
	docker build -t snip .

docker-run: ## Run with docker-compose
	docker-compose up -d

install: build ## Install binary to /usr/local/bin
	sudo cp $(BINARY) /usr/local/bin/snip

cross: ## Cross-compile for all platforms
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/snip-linux-amd64   $(MAIN)
	GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/snip-linux-arm64   $(MAIN)
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/snip-darwin-amd64  $(MAIN)
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/snip-darwin-arm64  $(MAIN)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/snip-windows-amd64.exe $(MAIN)

version: ## Show version
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"

deploy: cross ## Build and deploy to VPS
	scp dist/snip-linux-amd64 root@38.55.146.183:/opt/snip/snip
	ssh root@38.55.146.183 'killall -9 snip 2>/dev/null; sleep 1; nohup /opt/snip/snip > /opt/snip/snip.log 2>&1 &'
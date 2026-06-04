.PHONY: build run clean test docker

build:
	go build -ldflags="-s -w" -o bin/snip ./cmd/server

run: build
	./bin/snip

clean:
	rm -rf bin/ data/

test:
	go test ./... -v

docker:
	docker build -t snip .

docker-run:
	docker-compose up -d

fmt:
	go fmt ./...

lint:
	go vet ./...

all: lint sec test build run

build:
	@echo "Building..."
	@go build -o bin/main cmd/api/main.go 

run:
	@echo "Running..."
	@go run cmd/api/main.go

test:
	@echo "Testing..."
	@go clean -testcache
	@go test ./...


lint:
	@golangci-lint run -v

scan:
	@gosec -r


.PHONY: all build run test sec lint

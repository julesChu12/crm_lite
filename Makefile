.PHONY: lint test build swag ci

lint:
	golangci-lint run ./...

test:
	go test ./... -cover -count=1

build:
	go build ./...

swag:
	swag init -g main.go -o docs

ci: lint test build



APP_NAME=wk2mmd

.PHONY: all build lint test coverage

all: build

build:
	go build -o $(APP_NAME) .

lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo 'golangci-lint not found. Installing...'; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2)
	golangci-lint run ./...

test:
	cd $(CURDIR) && go test ./...

test-integration:
	go test ./internal/diagram/flowchart_integration_test.go

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
.PHONY: build clean test proto install deps tidy gen-docs

# Build variables
BINARY_NAME=it2
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
all: deps proto build

# Build the application
build: install

# Install the application
install:
	go install

clean:

# Generate protobuf code using Buf
proto:
	go generate ./...

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	go vet ./...
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Install tool dependencies
deps:
	go mod tidy
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Update dependencies
tidy:
	go mod tidy

# Create distribution binaries
dist: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_arm64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_windows_amd64.exe

# Generate documentation
gen-docs:
	rm ./docs/it2_*.md || true
	go run cmd/gen-docs/main.go --website --hierarchy --doc-path ./docs

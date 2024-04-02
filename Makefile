.PHONY: all build run clean fmt test

# Variables
PROJECT_NAME="chatApp"
GO_FILES=$(wildcard *.go)

# Default target
all: build

# Build the Go binary
build:
	go build -o bin/$(PROJECT_NAME) .

# Run the Go binary
run: build
	./bin/$(PROJECT_NAME)

# Clean up the binary
clean:
	rm -f bin/$(PROJECT_NAME)

# Format Go code
fmt:
	go fmt $(GO_FILES)

# Run tests
test:
	go test ./...
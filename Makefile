.PHONY: all build run clean fmt test test_coverage vet

# Variables
PROJECT_NAME="chat_app"
SRC_DIRS := cmd internal pkg
## Find all files in SRC_DIRS and its subdirs
GO_FILES := $(wildcard $(addsuffix /*.go, $(SRC_DIRS)) $(addsuffix /**/*.go, $(SRC_DIRS)) )


# Default target
all: build

# Build the Go binary
build: fmt
	go build -o bin/$(PROJECT_NAME) ./cmd

# Run the Go binary
run: build
	./bin/$(PROJECT_NAME)

# Clean up the binary
clean:
	go clean 
	rm -f bin/$(PROJECT_NAME)

# Format Go code
fmt:
	for f in $(GO_FILES); do gofmt -w $$f; done

# Run tests
test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

vet:
	go vet
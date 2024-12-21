.PHONY: all build test clean lint coverage install uninstall help

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=baler
BUILD_DIR=bin

LDFLAGS=-ldflags "-s -w"

GOLINT=golangci-lint

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/baler

build-all:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 ./cmd/baler
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 ./cmd/baler
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 ./cmd/baler
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 ./cmd/baler
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe ./cmd/baler
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_arm64.exe ./cmd/baler

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

lint:
	$(GOLINT) run

coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

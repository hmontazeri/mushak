.PHONY: build install clean test fmt vet

# Binary name
BINARY=mushak
VERSION?=dev

# Build variables
LDFLAGS=-ldflags "-X github.com/hmontazeri/mushak/pkg/version.Version=$(VERSION)"

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/mushak

# Install the binary to /usr/local/bin
install: build
	sudo mv $(BINARY) /usr/local/bin/

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf dist/

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet test

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 ./cmd/mushak
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 ./cmd/mushak
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 ./cmd/mushak
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./cmd/mushak

# Development build and run
dev: build
	./$(BINARY)

# Show help
help:
	@echo "Mushak - Build Commands"
	@echo ""
	@echo "make build      - Build the binary"
	@echo "make install    - Build and install to /usr/local/bin"
	@echo "make clean      - Remove build artifacts"
	@echo "make test       - Run tests"
	@echo "make fmt        - Format code"
	@echo "make vet        - Run go vet"
	@echo "make check      - Run fmt, vet, and test"
	@echo "make build-all  - Build for all platforms"
	@echo "make dev        - Build and run"

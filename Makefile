APP_NAME := proxy_app
GO := go
PKG := ./...
BIN_DIR := bin
BINARY := $(BIN_DIR)/$(APP_NAME)


.PHONY: all build test clean run fmt install

all: build

# Build the Go binary
build: clean install
	@echo "Building the application..."
	$(GO) build -o $(BINARY) .

# Run the Go application
run: build
	@echo "Running the application..."
	$(BINARY)

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v $(PKG)

# Clean up build files
clean:
	@echo "Cleaning up..."
	rm -rf $(BIN_DIR)

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(GO) mod tidy

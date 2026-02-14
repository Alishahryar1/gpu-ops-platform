# GPU Ops Platform Makefile

.PHONY: all build clean test install daemon cli run-daemon run-cli help

# Variables
BINARY_DIR = bin
DAEMON_BINARY = $(BINARY_DIR)/gputld
CLI_BINARY = $(BINARY_DIR)/gputl
GO = go

# Build targets
all: build

build: daemon cli

daemon:
	@echo "Building gputld daemon..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(DAEMON_BINARY) ./cmd/gputld
	@echo "Daemon built: $(DAEMON_BINARY)"

cli:
	@echo "Building gputl CLI..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(CLI_BINARY) ./cmd/gputl
	@echo "CLI built: $(CLI_BINARY)"

test:
	@echo "Running tests..."
	$(GO) test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf $(BINARY_DIR)
	$(GO) clean

install: build
	@echo "Installing..."
	install -m 755 $(DAEMON_BINARY) /usr/local/bin/gputld
	install -m 755 $(CLI_BINARY) /usr/local/bin/gputl
	@echo "Installed to /usr/local/bin"

run-daemon:
	@echo "Starting daemon..."
	$(DAEMON_BINARY) || $(GO) run ./cmd/gputld

run-cli:
	@echo "Running CLI..."
	$(CLI_BINARY) $(ARGS) || $(GO) run ./cmd/gputl $(ARGS)

# Policy engine (using UV)
policy-test:
	@echo "Testing policy engine..."
	cd python && uv sync && python -m starlark_engine.engine --load-all

policy-install:
	@echo "Installing policy engine dependencies with UV..."
	@which uv > /dev/null 2>&1 || (echo "UV not found. Install: pip install uv" && exit 1)
	cd python && uv sync

# Development
dev-deps:
	@echo "Installing Go dependencies..."
	$(GO) mod download

dev-setup: dev-deps policy-install
	@echo "Setting up development environment..."
	@mkdir -p /var/lib/gputl
	@mkdir -p /etc/gputl/policies
	@echo "Development environment ready"

# Docker (optional)
docker-build:
	@echo "Building Docker image..."
	docker build -t gpu-ops-platform:latest .

docker-run:
	@echo "Running Docker container..."
	docker run --gpus all -p 8080:8080 -p 9090:9090 gpu-ops-platform:latest

help:
	@echo "GPU Ops Platform - Available targets:"
	@echo ""
	@echo "  make build           - Build daemon and CLI binaries"
	@echo "  make daemon          - Build only the daemon"
	@echo "  make cli             - Build only the CLI"
	@echo "  make test            - Run tests"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make install         - Install to /usr/local/bin"
	@echo "  make run-daemon      - Run the daemon"
	@echo "  make run-cli         - Run the CLI (use ARGS='command')"
	@echo "  make policy-test     - Test the policy engine"
	@echo "  make policy-install  - Install policy engine dependencies"
	@echo "  make dev-deps        - Install Go dependencies"
	@echo "  make dev-setup       - Full development environment setup"
	@echo "  make help            - Show this help message"

# GPU Ops Platform

A GPU operations platform for data center GPU health monitoring, registration, and optimization.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     GPU Ops Platform                            │
├─────────────────────────────────────────────────────────────────┤
│  gputl CLI (Go) │ gputld Daemon (Go) │ Starlark Engine (Python) │
└─────────────────────────────────────────────────────────────────┘
```

## Components

| Component | Language | Description |
|-----------|----------|-------------|
| **gputld** | Go | Main daemon for GPU registration and health monitoring |
| **gputl** | Go | CLI tool for GPU operations |
| **Starlark** | Python | Policy engine for GPU allocation and optimization |

## Getting Started

### Prerequisites

- Go 1.21+
- Python 3.10+
- UV for Python dependency management (recommended)
  ```bash
  pip install uv
  ```
- NVIDIA GPU with NVML support

### Build

```bash
go mod download
go build -o bin/gputl ./cmd/gputl
go build -o bin/gputld ./cmd/gputld
```

### Run

```bash
# Start the daemon
./bin/gputld start

# Check GPU status
./bin/gputl status

# Register a new GPU
./bin/gputl register --id 0 --name "GPU0"
```

## Development

### Project Structure

```
├── cmd/
│   ├── gputld/          # Go daemon (core service)
│   └── gputl/           # CLI tool
├── pkg/
│   ├── gpu/             # GPU discovery and monitoring
│   ├── registration/    # GPU registration service
│   ├── health/          # Health checks
│   └── metrics/         # Prometheus metrics
├── policies/            # Starlark policy definitions
└── python/              # Starlark policy engine
```

## License

MIT

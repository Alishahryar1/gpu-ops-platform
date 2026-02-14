# GPU Ops Platform - Quick Start Guide

This guide will help you get the GPU Ops Platform up and running on your system.

## Prerequisites

- Go 1.21+ installed
- Python 3.10+ installed (for policy engine)
- UV installed (recommended for Python dependency management)
  ```bash
  pip install uv
  ```
- NVIDIA GPU with NVML support (for GPU monitoring)

## Installation

### Windows (PowerShell)

```powershell
# Navigate to project directory
cd C:\Users\alish\Desktop\projects\gpu-ops-platform

# Download dependencies
go mod download

# Build everything
.\build.ps1

# Or build and install to PATH
.\build.ps1 -Install
```

### Linux/macOS

```bash
# Navigate to project directory
cd gpu-ops-platform

# Download dependencies
go mod download

# Build everything
make build

# Or install to /usr/local/bin
make install
```

## Running the Platform

### Start the Daemon

```powershell
# Windows PowerShell
.\build.ps1 -RunDaemon

# Or if binaries are already built
.\bin\gputld.exe

# Linux/macOS
make run-daemon
# or
./bin/gputld
```

The daemon will start:
- HTTP API server on `http://localhost:8080`
- Prometheus metrics on `http://localhost:9090/metrics`
- Health monitoring loop

### Use the CLI

```powershell
# Check daemon health
.\bin\gputl.exe health

# List all GPUs
.\bin\gputl.exe status

# Get status of specific GPU
.\bin\gputl.exe status 0

# View health check configurations
.\bin\gputl.exe health-checks

# Register a GPU
.\bin\gputl.exe register 0 --name "RTX_5070_Ti" --pool "production" --tags "high-perf,ml"
```

### Test Policy Engine

```powershell
# Install UV (if not already installed)
pip install uv

# Install Python dependencies with UV
cd python/starlark_engine
uv sync
# or
uv pip install -r requirements.txt

# Load and test policies
python engine.py --load-all

# View registered pools
python engine.py --load-all --list-pools

# Test allocation for specific GPU
python engine.py --load-all --test-gpu 0
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/v1/gpus` | GET | List all GPUs |
| `/api/v1/gpus/:id` | GET | Get GPU details |
| `/api/v1/register` | POST | Register a GPU |
| `/api/v1/unregister/:id` | DELETE | Unregister a GPU |
| `/api/v1/healthchecks` | GET | List health check configurations |

### Example API Usage

```bash
# List all GPUs
curl http://localhost:8080/api/v1/gpus

# Health check
curl http://localhost:8080/health

# Prometheus metrics
curl http://localhost:9090/metrics
```

## Project Structure

```
gpu-ops-platform/
├── cmd/
│   ├── gputld/          # Daemon entry point
│   └── gputl/           # CLI entry point
├── pkg/
│   ├── gpu/             # GPU discovery and monitoring
│   ├── registration/    # GPU registration service
│   ├── health/          # Health checks
│   ├── metrics/         # Prometheus metrics
│   └── config/          # Configuration management
├── policies/            # Starlark policy definitions
│   ├── production.gsky  # Production pool policy
│   └── development.gsky # Development pool policy
├── python/
│   └── starlark_engine/ # Starlark policy evaluator
├── bin/                 # Built binaries (created after build)
├── build.ps1            # Windows build script
├── Makefile             # Unix Makefile
├── go.mod               # Go module dependencies
└── README.md            # Main documentation
```

## Next Steps

1. **Explore the code**: Start by looking at `pkg/gpu/gpu.go` for GPU discovery
2. **Write a policy**: Create a new `.gsky` file in `policies/`
3. **Add health checks**: See `pkg/health/health.go` to understand health monitoring
4. **NVML Integration**: Replace mock GPU data with actual `github.com/NVIDIA/go-nvml` calls
5. **Add web UI**: Build a dashboard to visualize GPU status and health

## Troubleshooting

### "go: command not found"
Go is not in your PATH. Install Go from https://go.dev/dl/ and add it to your PATH.

### "Python not found" when testing policies
Install Python 3.10+ from https://www.python.org/downloads/ and run:
```powershell
# Install UV first
pip install uv

# Install dependencies with UV
cd python/starlark_engine
uv sync
```

### Port already in use (8080 or 9090)
Kill the process using the port or change the port in the code.

## Development Tips

- Run `go generate` if you add new protobuf files
- Use `go test -v ./...` to run all tests
- Check `/metrics` endpoint for current GPU metrics
- Look at `policies/` for examples of Starlark policy syntax
- Use `uv sync` in `python/starlark_engine/` to install dependencies fast
- The Starlark engine uses pure Python for evaluation (no Go integration yet)

## License

MIT

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

## Prerequisites

- Go 1.21+
- Python 3.10+ (for policy engine)
- UV for Python dependency management (recommended)
  ```bash
  pip install uv
  ```
- NVIDIA GPU with NVML support

## Installation

### Windows (PowerShell)

```powershell
cd gpu-ops-platform
go mod download
.\build.ps1

# Or build and install to PATH
.\build.ps1 -Install
```

### Linux/macOS

```bash
cd gpu-ops-platform
go mod download
make build

# Or install to /usr/local/bin
make install
```

## Running the Platform

### Start the Daemon

```powershell
# Windows
.\build.ps1 -RunDaemon
# Or if binaries are already built
.\bin\gputld.exe
```

```bash
# Linux/macOS
make run-daemon
# or
./bin/gputld
```

The daemon starts:
- HTTP API server on `http://localhost:8080`
- Prometheus metrics on `http://localhost:9090/metrics`
- Health monitoring loop

### Use the CLI

```powershell
# Windows
.\bin\gputl.exe health
.\bin\gputl.exe status
.\bin\gputl.exe status 0
.\bin\gputl.exe health-checks
.\bin\gputl.exe register 0 --name "RTX_5070_Ti" --pool "production" --tags "high-perf,ml"
```

```bash
# Linux/macOS
./bin/gputl health
./bin/gputl status
./bin/gputl status 0
./bin/gputl health-checks
./bin/gputl register 0 --name "RTX_5070_Ti" --pool "production" --tags "high-perf,ml"
```

### Test Policy Engine

```bash
pip install uv
cd python
uv sync

# Load and test policies
uv run star --load-all

# View registered pools
uv run star --load-all --list-pools

# Test allocation for specific GPU
uv run star --load-all --test-gpu 0
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
curl http://localhost:8080/api/v1/gpus
curl http://localhost:8080/health
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
└── go.mod               # Go module dependencies
```

## Development

- Run `go generate` if you add new protobuf files
- Use `go test -v ./...` to run all tests
- Check `/metrics` endpoint for current GPU metrics
- Look at `policies/` for examples of Starlark policy syntax
- Use `uv sync` in `python/` to install dependencies
- The Starlark engine uses pure Python for evaluation (no Go integration yet)

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
```bash
pip install uv
cd python
uv sync
```

### Port already in use (8080 or 9090)
Kill the process using the port or change the port in the code.

## License

MIT

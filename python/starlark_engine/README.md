# Starlark Policy Engine for GPU Ops Platform

A Python-based Starlark policy engine for defining and evaluating GPU allocation, scheduling, and optimization policies. Uses `starlark-go` - the Python bindings for Go's Starlark interpreter (go.starlark.net).

## Requirements

- Python 3.10+
- [UV](https://github.com/astral-sh/uv) for dependency management

## Installation with UV

```bash
# From project root
cd python

# Install dependencies
uv sync
```

## Usage

### Command Line

Run from the `python` directory:

```bash
cd python

# Load and test all policies
uv run star --load-all --list-pools

# Test allocation for specific GPU
uv run star --load-all --test-gpu 0

# Load specific policy file
uv run star --load policies/development.gsky
```

Or use the build script from project root:

```powershell
# Windows
.\build.ps1 -PolicyTest

# Linux/macOS
make policy-test
```

### Python API

```python
from starlark_engine import PolicyEngine, GPUInfo

# Create engine
engine = PolicyEngine(policy_dir="policies")

# Load all policies
engine.load_all_policies()

# Define a GPU
gpu = GPUInfo(
    id=0,
    name="NVIDIA GeForce RTX 5070 Ti",
    uuid="GPU-00000000",
    memory_gb=24.0,
    temperature_c=35.0,
    power_w=10.0,
    tags=["high-perf"]
)

# Evaluate allocation
result = engine.evaluate_allocation(
    gpu=gpu,
    requirements={"min_memory": 16, "tags": ["high-perf"]}
)
print(f"Allocation result: {result}")
```

## Policy Syntax

Policies are written in Starlark (`.gsky` files):

```python
# Define a GPU pool
def production_pool():
    return gpu_pool(
        name = "gpu_pool_prod",
        gpu_type = ["RTX_5070ti", "A100"],
        min_memory_gb = 16,
        max_temp_c = 85,
        power_policy = "performance"
    )

# Register the pool
gpu_ops.register_pool(production_pool())
```

## About starlark-go

This package uses `starlark-go` - Python bindings for go.starlark.net, which is the same Starlark implementation used internally at NVIDIA for GPU resource management and configuration.

## Development

```bash
cd python

# Install dev dependencies
uv sync --dev

# Run tests
pytest

# Lint code
ruff check .
```

# Starlark Policy Engine for GPU Ops Platform

A Python-based Starlark policy engine for defining and evaluating GPU allocation, scheduling, and optimization policies.

## Requirements

- Python 3.10+
- [UV](https://github.com/astral-sh/uv) for dependency management

## Installation with UV

```bash
# Create virtual environment and install dependencies
uv venv

# Activate venv (Windows)
.venv\Scripts\activate

# Activate venv (Linux/macOS)
source .venv/bin/activate

# Install dependencies
uv pip install -r requirements.txt
# OR simply:
uv sync
```

## Usage

### Command Line

```bash
# Load and test all policies
python engine.py --load-all --list-pools

# Test allocation for specific GPU
python engine.py --load-all --test-gpu 0

# Load specific policy file
python engine.py --load development.gsky
```

### Python API

```python
from starlark_engine import PolicyEngine, GPUInfo

# Create engine
engine = PolicyEngine(policy_dir="../../policies")

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

## Development

```bash
# Run tests (once dependencies are installed)
pytest

# Lint code
ruff check .
```

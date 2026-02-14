"""
Starlark Policy Engine for GPU Ops Platform

This module provides a Starlark-based policy engine for defining and evaluating
GPU allocation, scheduling, and optimization policies.
"""

import os
import sys
from typing import Any, Dict, List, Optional
from pathlib import Path
from dataclasses import dataclass, field

try:
    from starlark import Starlark
except ImportError:
    # Fallback to go.starlark.net via Python bindings
    Starlark = None

# ============================================================================
# Policy Data Models
# ============================================================================

@dataclass
class GPUInfo:
    """Represents a single GPU device."""
    id: int
    name: str
    uuid: str
    memory_gb: float
    temperature_c: float
    power_w: float
    online: bool = True
    registered: bool = False
    pool: str = ""
    tags: List[str] = field(default_factory=list)


@dataclass
class GPUPool:
    """Defines a GPU pool configuration."""
    name: str
    gpu_types: List[str] = field(default_factory=list)
    min_memory_gb: float = 0.0
    max_temp_c: float = 100.0
    power_policy: str = "adaptive"
    health_threshold: float = 0.9
    opt_strategies: List[Dict[str, Any]] = field(default_factory=list)
    health_checks: List[Dict[str, Any]] = field(default_factory=list)


@dataclass
class ScheduleRule:
    """Defines a scheduling rule."""
    type: str
    metric: Optional[str] = None
    strategy: Optional[str] = None
    values: Optional[List[str]] = None
    allow: Optional[bool] = None


@dataclass
class ScheduleRuleset:
    """Defines a set of scheduling rules."""
    name: str
    rules: List[ScheduleRule] = field(default_factory=list)


@dataclass
class Policy:
    """Represents a complete policy document."""
    name: str
    version: str = "1.0"
    pools: List[GPUPool] = field(default_factory=list)
    schedules: List[ScheduleRuleset] = field(default_factory=list)


# ============================================================================
# Starlark Built-in Functions
# ============================================================================

class GPUOpsModule:
    """Provides GPU operations for Starlark policies."""

    def __init__(self):
        self.pools: Dict[str, GPUPool] = {}
        self.schedules: Dict[str, ScheduleRuleset] = {}

    def register_pool(self, pool: GPUPool) -> None:
        """Register a GPU pool."""
        self.pools[pool.name] = pool
        print(f"[Policy] Registered pool: {pool.name}")

    def register_schedule(self, schedule: ScheduleRuleset) -> None:
        """Register a schedule ruleset."""
        self.schedules[schedule.name] = schedule
        print(f"[Policy] Registered schedule: {schedule.name}")

    def get_pool(self, name: str) -> Optional[GPUPool]:
        """Get a pool by name."""
        return self.pools.get(name)

    def get_all_pools(self) -> List[GPUPool]:
        """Get all pools."""
        return list(self.pools.values())

    def get_schedule(self, name: str) -> Optional[ScheduleRuleset]:
        """Get a schedule by name."""
        return self.schedules.get(name)


# ============================================================================
# Policy Engine
# ============================================================================

class PolicyEngine:
    """Main policy engine for evaluating Starlark policies."""

    def __init__(self, policy_dir: Optional[str] = None):
        self.policy_dir = Path(policy_dir) if policy_dir else Path("./policies")
        self.policies: Dict[str, Policy] = {}
        self.gpu_ops = GPUOpsModule()
        self._setup_starlark_globals()

    def _setup_starlark_globals(self):
        """Set up global symbols available to Starlark scripts."""
        self.globals = {
            # GPU pool functions
            "gpu_pool": self._create_gpu_pool,
            "optimize_clock": self._create_optimize_clock,
            "optimize_memory": self._create_optimize_memory,
            "optimize_power": self._create_optimize_power,
            "check": self._create_health_check,

            # Scheduling functions
            "priority_rule": self._create_priority_rule,
            "balance_rule": self._create_balance_rule,
            "distribution_rule": self._create_distribution_rule,
            "preemption_rule": self._create_preemption_rule,

            # Registration
            "schedule_ruleset": self._create_schedule_ruleset,
            "gpu_ops": self.gpu_ops,
        }

    # ---------------------------------------------------------------------
    # Starlark Factory Functions
    # ---------------------------------------------------------------------

    def _create_gpu_pool(self, **kwargs) -> GPUPool:
        """Create a GPU pool from kwargs."""
        # Extract type hints
        gpu_type = kwargs.get('gpu_type', [])
        if isinstance(gpu_type, str):
            gpu_type = [gpu_type]

        return GPUPool(
            name=kwargs.get('name', 'default'),
            gpu_types=list(gpu_type),
            min_memory_gb=kwargs.get('min_memory_gb', 0.0),
            max_temp_c=kwargs.get('max_temp_c', 100.0),
            power_policy=kwargs.get('power_policy', 'adaptive'),
            health_threshold=kwargs.get('health_threshold', 0.9),
            opt_strategies=kwargs.get('opt_strategies', []),
            health_checks=kwargs.get('health_checks', []),
        )

    def _create_optimize_clock(self, **kwargs) -> Dict[str, Any]:
        return {"type": "clock", **kwargs}

    def _create_optimize_memory(self, **kwargs) -> Dict[str, Any]:
        return {"type": "memory", **kwargs}

    def _create_optimize_power(self, **kwargs) -> Dict[str, Any]:
        return {"type": "power", **kwargs}

    def _create_health_check(self, name: str, **kwargs) -> Dict[str, Any]:
        return {"name": name, **kwargs}

    def _create_priority_rule(self, **kwargs) -> ScheduleRule:
        return ScheduleRule(type="priority", **kwargs)

    def _create_balance_rule(self, **kwargs) -> ScheduleRule:
        return ScheduleRule(type="balance", **kwargs)

    def _create_distribution_rule(self, **kwargs) -> ScheduleRule:
        return ScheduleRule(type="distribution", **kwargs)

    def _create_preemption_rule(self, **kwargs) -> ScheduleRule:
        return ScheduleRule(type="preemption", **kwargs)

    def _create_schedule_ruleset(self, **kwargs) -> ScheduleRuleset:
        return ScheduleRuleset(**kwargs)

    # ---------------------------------------------------------------------
    # Policy Loading
    # ---------------------------------------------------------------------

    def load_policy(self, policy_path: str) -> bool:
        """Load a policy from a file."""
        path = Path(policy_path)
        if not path.exists():
            # Try relative to policy_dir
            path = self.policy_dir / policy_path
            if not path.exists():
                print(f"[Policy] Policy not found: {policy_path}")
                return False

        print(f"[Policy] Loading policy from: {path}")

        try:
            # Read policy file
            with open(path, 'r') as f:
                code = f.read()

            # For now, just parse and show what would be loaded
            # In production, use actual Starlark interpreter
            self._parse_policy_simple(code, path.stem)

            return True

        except Exception as e:
            print(f"[Policy] Error loading policy: {e}")
            return False

    def load_all_policies(self) -> int:
        """Load all policies from the policy directory."""
        count = 0
        if not self.policy_dir.exists():
            print(f"[Policy] Policy directory not found: {self.policy_dir}")
            return count

        for policy_file in self.policy_dir.glob("*.gsky"):
            if self.load_policy(policy_file):
                count += 1

        print(f"[Policy] Loaded {count} policy file(s)")
        return count

    def _parse_policy_simple(self, code: str, name: str):
        """
        Simple policy parser (placeholder for actual Starlark interpreter).
        In production, this would use the go.starlark.net or pure Python Starlark.
        """
        # Extract basic information from the policy
        print(f"[Policy] Parsing: {name}")
        print(f"[Policy] This is a placeholder for actual Starlark execution")

        # For now, create a mock policy with basic info
        self.policies[name] = Policy(name=name)

        # Extract pool definitions (simple regex-based for demo)
        import re
        pool_names = re.findall(r'name\s*=\s*["\'](\w+)["\']', code)
        for pool_name in pool_names:
            print(f"[Policy] Detected pool: {pool_name}")

    # ---------------------------------------------------------------------
    # Policy Evaluation
    # ---------------------------------------------------------------------

    def evaluate_allocation(self, gpu: GPUInfo, requirements: Dict[str, Any]) -> bool:
        """Evaluate if a GPU meets allocation requirements."""
        # Get all pools
        pools = self.gpu_ops.get_all_pools()

        for pool in pools:
            if self._matches_pool(gpu, pool) and self._meets_requirements(gpu, requirements):
                return True

        return False

    def _matches_pool(self, gpu: GPUInfo, pool: GPUPool) -> bool:
        """Check if GPU matches pool criteria."""
        if pool.gpu_types and gpu.name not in pool.gpu_types:
            # Partial matching (contains any of the types)
            if not any(t in gpu.name for t in pool.gpu_types):
                return False

        if gpu.memory_gb < pool.min_memory_gb:
            return False

        if gpu.temperature_c > pool.max_temp_c:
            return False

        if gpu.pool and gpu.pool != pool.name:
            return False

        return True

    def _meets_requirements(self, gpu: GPUInfo, requirements: Dict[str, Any]) -> bool:
        """Check if GPU meets specific requirements."""
        if 'min_memory' in requirements and gpu.memory_gb < requirements['min_memory']:
            return False

        if 'max_temp' in requirements and gpu.temperature_c > requirements['max_temp']:
            return False

        if 'tags' in requirements:
            required_tags = set(requirements['tags'])
            available_tags = set(gpu.tags)
            if not required_tags.issubset(available_tags):
                return False

        return True

    def get_recommended_gpus(self, gpus: List[GPUInfo], requirements: Dict[str, Any]) -> List[GPUInfo]:
        """Get list of recommended GPUs based on policy."""
        recommended = []

        for gpu in gpus:
            if self.evaluate_allocation(gpu, requirements):
                score = self._score_gpu(gpu)
                recommended.append((gpu, score))

        # Sort by score (descending)
        recommended.sort(key=lambda x: x[1], reverse=True)
        return [gpu for gpu, _ in recommended]

    def _score_gpu(self, gpu: GPUInfo) -> float:
        """Calculate a suitability score for a GPU."""
        score = 1.0

        # Prefer GPUs with better health
        if gpu.online:
            score += 0.5

        # Prefer GPUs with more memory
        if gpu.memory_gb >= 16:
            score += 0.3

        # Prefer cooler GPUs
        if gpu.temperature_c < 60:
            score += 0.2

        # Penalize high temperature
        elif gpu.temperature_c > 80:
            score -= 0.3

        return score


# ============================================================================
# CLI Interface
# ============================================================================

def main():
    """CLI for the policy engine."""
    import argparse

    parser = argparse.ArgumentParser(description="GPU Ops Policy Engine")
    parser.add_argument("--policy-dir", default="./policies", help="Policy directory")
    parser.add_argument("--load", help="Load specific policy file")
    parser.add_argument("--load-all", action="store_true", help="Load all policies")
    parser.add_argument("--list-pools", action="store_true", help="List registered pools")
    parser.add_argument("--list-schedules", action="store_true", help="List registered schedules")
    parser.add_argument("--test-gpu", type=int, help="Test allocation for specific GPU")

    args = parser.parse_args()

    engine = PolicyEngine(args.policy_dir)

    if args.load:
        engine.load_policy(args.load)

    if args.load_all:
        engine.load_all_policies()

    if args.list_pools:
        print("\nRegistered Pools:")
        for pool in engine.gpu_ops.get_all_pools():
            print(f"  - {pool.name}: {pool.gpu_types if pool.gpu_types else 'any type'}")

    if args.list_schedules:
        print("\nRegistered Schedules:")
        for name in engine.gpu_ops.schedules.keys():
            print(f"  - {name}")

    if args.test_gpu is not None:
        test_gpu = GPUInfo(
            id=args.test_gpu,
            name="NVIDIA GeForce RTX 5070 Ti",
            uuid=f"GPU-{args.test_gpu:08x}",
            memory_gb=24.0,
            temperature_c=35.0,
            power_w=10.0
        )
        result = engine.evaluate_allocation(test_gpu, {"min_memory": 16})
        print(f"\nGPU {args.test_gpu} allocation result: {result}")


if __name__ == "__main__":
    main()

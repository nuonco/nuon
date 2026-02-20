"""Step primitives for the Cerberus DSL.

Each function returns a Step dataclass that describes what to do.
The runner interprets these into Temporal activities.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import timedelta
from typing import Any, Callable


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def days(n: int) -> timedelta:
    return timedelta(days=n)


def hours(n: int) -> timedelta:
    return timedelta(hours=n)


def minutes(n: int) -> timedelta:
    return timedelta(minutes=n)


# ---------------------------------------------------------------------------
# Base step type
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class Step:
    """Base class for all step primitives."""

    kind: str
    params: dict[str, Any] = field(default_factory=dict)

    def __repr__(self) -> str:
        parts = [f"{k}={v!r}" for k, v in self.params.items() if v is not None]
        return f"{self.kind}({', '.join(parts)})"


# ---------------------------------------------------------------------------
# Deploy & Build
# ---------------------------------------------------------------------------


def deploy(
    *,
    wait: bool = True,
    install: str | None = None,
    plan_only: bool = False,
) -> Step:
    """Trigger a deploy to one or all installs."""
    return Step(
        kind="deploy",
        params={"wait": wait, "install": install, "plan_only": plan_only},
    )


def build(*, component: str | None = None) -> Step:
    """Trigger a build for one or all components."""
    return Step(kind="build", params={"component": component})


def teardown(*, component: str | None = None) -> Step:
    """Teardown one or all components on all installs."""
    return Step(kind="teardown", params={"component": component})


# ---------------------------------------------------------------------------
# Workflow & Approval
# ---------------------------------------------------------------------------


def approve_workflow(*, workflow_id: str = "{last_workflow_id}") -> Step:
    return Step(kind="approve_workflow", params={"workflow_id": workflow_id})


def cancel_workflow(*, workflow_id: str = "{last_workflow_id}") -> Step:
    return Step(kind="cancel_workflow", params={"workflow_id": workflow_id})


def retry_workflow_step(*, workflow_id: str, step_id: str) -> Step:
    return Step(
        kind="retry_workflow_step",
        params={"workflow_id": workflow_id, "step_id": step_id},
    )


def get_workflows(*, install: str = "{install_id}") -> Step:
    return Step(kind="get_workflows", params={"install": install})


def wait_for_workflow(
    *,
    status: str,
    timeout: timedelta = timedelta(minutes=30),
) -> Step:
    return Step(
        kind="wait_for_workflow",
        params={"status": status, "timeout_seconds": int(timeout.total_seconds())},
    )


# ---------------------------------------------------------------------------
# Config & Component Changes
# ---------------------------------------------------------------------------


def update_component(component_name: str, **kwargs: Any) -> Step:
    return Step(
        kind="update_component",
        params={"component": component_name, **kwargs},
    )


def update_app_config(**kwargs: Any) -> Step:
    return Step(kind="update_app_config", params=kwargs)


def update_app_inputs(*, inputs: dict[str, Any]) -> Step:
    return Step(kind="update_app_inputs", params={"inputs": inputs})


def update_install_inputs(*, install: str, inputs: dict[str, Any]) -> Step:
    return Step(
        kind="update_install_inputs",
        params={"install": install, "inputs": inputs},
    )


def create_secret(*, name: str, value: str) -> Step:
    return Step(kind="create_secret", params={"name": name, "value": value})


def delete_secret(*, name: str) -> Step:
    return Step(kind="delete_secret", params={"name": name})


# ---------------------------------------------------------------------------
# Install Lifecycle
# ---------------------------------------------------------------------------


def reprovision(*, install: str | None = None) -> Step:
    return Step(kind="reprovision", params={"install": install})


def deprovision(*, install: str | None = None) -> Step:
    return Step(kind="deprovision", params={"install": install})


def create_install(*, name: str) -> Step:
    return Step(kind="create_install", params={"name": name})


def delete_install(*, install: str) -> Step:
    return Step(kind="delete_install", params={"install": install})


# ---------------------------------------------------------------------------
# Releases
# ---------------------------------------------------------------------------


def create_release(*, component: str) -> Step:
    return Step(kind="create_release", params={"component": component})


def get_releases(*, component: str) -> Step:
    return Step(kind="get_releases", params={"component": component})


# ---------------------------------------------------------------------------
# Verification & Assertions
# ---------------------------------------------------------------------------


def verify_status(expected: str, *, install: str | None = None) -> Step:
    return Step(
        kind="verify_status",
        params={"expected": expected, "install": install},
    )


def verify_deploy_status(expected: str) -> Step:
    return Step(kind="verify_deploy_status", params={"expected": expected})


def verify_build_status(expected: str, *, component: str | None = None) -> Step:
    return Step(
        kind="verify_build_status",
        params={"expected": expected, "component": component},
    )


def assert_expr(fn: Callable, *, msg: str = "") -> Step:
    return Step(kind="assert_expr", params={"fn": fn, "msg": msg})


def assert_api(
    method: str,
    path: str,
    *,
    status: int | None = None,
    json_path: str | None = None,
    equals: Any = None,
) -> Step:
    return Step(
        kind="assert_api",
        params={
            "method": method,
            "path": path,
            "status": status,
            "json_path": json_path,
            "equals": equals,
        },
    )


# ---------------------------------------------------------------------------
# Control Flow
# ---------------------------------------------------------------------------


def delay(
    *,
    seconds: int = 0,
    minutes: int = 0,
    hours: int = 0,
    days: int = 0,
) -> Step:
    total = timedelta(seconds=seconds, minutes=minutes, hours=hours, days=days)
    return Step(kind="delay", params={"seconds": int(total.total_seconds())})


def repeat(
    *,
    n: int,
    steps: list[Step],
    interval: timedelta | None = None,
) -> Step:
    return Step(
        kind="repeat",
        params={
            "n": n,
            "steps": steps,
            "interval_seconds": int(interval.total_seconds()) if interval else 0,
        },
    )


# ---------------------------------------------------------------------------
# Fault Injection
# ---------------------------------------------------------------------------


def fault(type: str, **kwargs: Any) -> Step:
    return Step(kind="fault", params={"type": type, **kwargs})


def clear_faults() -> Step:
    return Step(kind="clear_faults")


def clear_fault(*, type: str) -> Step:
    return Step(kind="clear_fault", params={"type": type})


# ---------------------------------------------------------------------------
# CLI, API & Shell
# ---------------------------------------------------------------------------


def cli_run(command: str) -> Step:
    return Step(kind="cli_run", params={"command": command})


def api_call(method: str, path: str, *, body: dict | None = None) -> Step:
    return Step(kind="api_call", params={"method": method, "path": path, "body": body})


def shell(command: str) -> Step:
    return Step(kind="shell", params={"command": command})


# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------


def log(message: str, *, data: dict[str, Any] | None = None) -> Step:
    return Step(kind="log", params={"message": message, "data": data})

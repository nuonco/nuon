"""Canary runner — interprets a CanaryDefinition into an executable plan.

The runner produces a flat list of ExecutableStep objects that the Temporal
workflow iterates through. Setup and cleanup phases are generated automatically
from the canary's App/Install declarations.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum

from cerberus.dsl.base import App, Install
from cerberus.dsl.primitives import Step
from cerberus.dsl.registry import CanaryDefinition


class Phase(str, Enum):
    SETUP = "setup"
    STEPS = "steps"
    CLEANUP = "cleanup"


@dataclass
class ExecutableStep:
    """A single step in the execution plan."""

    index: int
    phase: Phase
    name: str  # human-readable name for display
    step: Step  # the underlying primitive


def build_execution_plan(defn: CanaryDefinition) -> list[ExecutableStep]:
    """Convert a canary definition into a flat execution plan.

    The plan has three phases:
    1. SETUP: create org, connect VCS, add support users, create app, create installs
    2. STEPS: the user-declared steps (with repeat/delay expanded at runtime)
    3. CLEANUP: deprovision installs, delete app, delete org
    """
    plan: list[ExecutableStep] = []
    idx = 0

    # -----------------------------------------------------------------------
    # Setup phase
    # -----------------------------------------------------------------------

    setup_steps = _build_setup_steps(defn.app, defn.installs)
    for name, step in setup_steps:
        plan.append(ExecutableStep(index=idx, phase=Phase.SETUP, name=name, step=step))
        idx += 1

    # -----------------------------------------------------------------------
    # User-declared steps
    # -----------------------------------------------------------------------

    for step in _flatten_steps(defn.steps):
        plan.append(
            ExecutableStep(
                index=idx,
                phase=Phase.STEPS,
                name=repr(step),
                step=step,
            )
        )
        idx += 1

    # -----------------------------------------------------------------------
    # Cleanup phase
    # -----------------------------------------------------------------------

    cleanup_steps = _build_cleanup_steps(defn.installs)
    for name, step in cleanup_steps:
        plan.append(ExecutableStep(index=idx, phase=Phase.CLEANUP, name=name, step=step))
        idx += 1

    return plan


def _build_setup_steps(app: App, installs: list[Install]) -> list[tuple[str, Step]]:
    """Generate the setup phase steps."""
    steps: list[tuple[str, Step]] = []

    steps.append(("Create canary org", Step(kind="setup_create_org", params={"sandbox_mode": True})))
    steps.append(("Connect VCS", Step(kind="setup_connect_vcs", params={})))
    steps.append(("Add support users", Step(kind="setup_add_support_users", params={})))
    steps.append((
        f"Create app '{app.name}'",
        Step(kind="setup_create_app", params={"app": app}),
    ))

    for install in installs:
        steps.append((
            f"Create install '{install.name}'",
            Step(kind="setup_create_install", params={"install": install}),
        ))

    return steps


def _build_cleanup_steps(installs: list[Install]) -> list[tuple[str, Step]]:
    """Generate the cleanup phase steps."""
    steps: list[tuple[str, Step]] = []

    steps.append(("Clear faults", Step(kind="clear_faults", params={})))

    for install in installs:
        steps.append((
            f"Deprovision install '{install.name}'",
            Step(kind="cleanup_deprovision_install", params={"install_name": install.name}),
        ))

    steps.append(("Delete app", Step(kind="cleanup_delete_app", params={})))
    steps.append(("Delete org", Step(kind="cleanup_delete_org", params={})))

    return steps


def _flatten_steps(steps: list[Step]) -> list[Step]:
    """Flatten steps for the execution plan.

    Note: repeat() steps are NOT expanded here — they are expanded at runtime
    by the Temporal workflow, since they involve timers between iterations.
    All other steps pass through as-is.
    """
    return steps

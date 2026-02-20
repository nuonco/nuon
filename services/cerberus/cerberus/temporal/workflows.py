"""Temporal workflow for executing a canary run.

One workflow per canary run. The workflow:
1. Executes setup steps (create org, app, installs)
2. Executes user-declared steps (deploy, verify, delay, repeat, etc.)
3. Executes cleanup steps (always, even on failure)
"""

from __future__ import annotations

from datetime import timedelta
from typing import Any

from temporalio import workflow

with workflow.unsafe.imports_passed_through():
    from cerberus.dsl.primitives import Step
    from cerberus.dsl.registry import get_canary
    from cerberus.dsl.runner import ExecutableStep, Phase, build_execution_plan
    from cerberus.temporal.activities import (
        execute_step,
        save_run_context,
        update_run_status,
    )


TASK_QUEUE = "cerberus"


@workflow.defn
class CanaryRunWorkflow:
    """Execute a complete canary run as a Temporal workflow."""

    def __init__(self) -> None:
        self._cancel_requested = False
        self._injected_faults: list[dict[str, Any]] = []

    @workflow.signal
    async def cancel_run(self) -> None:
        self._cancel_requested = True

    @workflow.signal
    async def inject_fault(self, fault_config: dict[str, Any]) -> None:
        self._injected_faults.append(fault_config)

    @workflow.run
    async def run(
        self,
        run_id: str,
        canary_name: str,
        params: dict[str, Any],
        nuon_base_url: str,
        nuon_api_token: str,
    ) -> dict[str, Any]:
        # Mark run as running
        await workflow.execute_activity(
            update_run_status,
            args=[run_id, "running"],
            start_to_close_timeout=timedelta(seconds=30),
        )

        # Build execution plan from canary definition
        defn = get_canary(canary_name)
        plan = build_execution_plan(defn)

        context_data: dict[str, Any] = {}
        failed = False
        failure_step: str | None = None
        failure_message: str | None = None

        # ---------------------------------------------------------------
        # Execute setup + steps phases
        # ---------------------------------------------------------------

        for exe_step in plan:
            if exe_step.phase == Phase.CLEANUP:
                continue  # cleanup runs after

            if self._cancel_requested:
                failed = True
                failure_message = "Cancelled by user"
                break

            if failed:
                break  # skip remaining steps on failure

            try:
                context_data = await self._execute_one(
                    run_id, exe_step, context_data, nuon_base_url, nuon_api_token
                )

                # Persist context after setup steps (so we track resources for cleanup)
                if exe_step.phase == Phase.SETUP:
                    await workflow.execute_activity(
                        save_run_context,
                        args=[run_id, context_data],
                        start_to_close_timeout=timedelta(seconds=30),
                    )

            except Exception as e:
                failed = True
                failure_step = exe_step.name
                failure_message = str(e)

        # ---------------------------------------------------------------
        # Cleanup phase (always runs)
        # ---------------------------------------------------------------

        for exe_step in plan:
            if exe_step.phase != Phase.CLEANUP:
                continue

            try:
                context_data = await self._execute_one(
                    run_id, exe_step, context_data, nuon_base_url, nuon_api_token
                )
            except Exception as e:
                # Log cleanup errors but don't change the run result
                workflow.logger.warning(f"Cleanup step '{exe_step.name}' failed: {e}")

        # ---------------------------------------------------------------
        # Final status
        # ---------------------------------------------------------------

        final_status = "failed" if failed else "passed"
        await workflow.execute_activity(
            update_run_status,
            args=[run_id, final_status, failure_message, failure_step],
            start_to_close_timeout=timedelta(seconds=30),
        )

        return {
            "status": final_status,
            "error_message": failure_message,
            "error_step": failure_step,
        }

    async def _execute_one(
        self,
        run_id: str,
        exe_step: ExecutableStep,
        context_data: dict[str, Any],
        nuon_base_url: str,
        nuon_api_token: str,
    ) -> dict[str, Any]:
        """Execute a single step, handling delay and repeat specially."""
        step = exe_step.step

        # delay() → use Temporal timer (no activity needed)
        if step.kind == "delay":
            seconds = step.params.get("seconds", 0)
            if seconds > 0:
                await workflow.sleep(timedelta(seconds=seconds))
            return context_data

        # repeat() → loop with optional interval
        if step.kind == "repeat":
            return await self._execute_repeat(
                run_id, exe_step, context_data, nuon_base_url, nuon_api_token
            )

        # Everything else → execute as activity
        return await workflow.execute_activity(
            execute_step,
            args=[
                run_id,
                exe_step.index,
                step.kind,
                self._serialize_params(step.params),
                exe_step.name,
                exe_step.phase.value,
                context_data,
                nuon_base_url,
                nuon_api_token,
            ],
            start_to_close_timeout=timedelta(minutes=30),
            retry_policy=workflow.RetryPolicy(maximum_attempts=1),
        )

    async def _execute_repeat(
        self,
        run_id: str,
        exe_step: ExecutableStep,
        context_data: dict[str, Any],
        nuon_base_url: str,
        nuon_api_token: str,
    ) -> dict[str, Any]:
        """Handle repeat() by iterating with optional sleep intervals."""
        params = exe_step.step.params
        n = params["n"]
        inner_steps: list[Step] = params["steps"]
        interval_seconds = params.get("interval_seconds", 0)

        for i in range(1, n + 1):
            if self._cancel_requested:
                break

            # Update iteration in context
            context_data["iteration"] = i

            for inner_step in inner_steps:
                if self._cancel_requested:
                    break

                if inner_step.kind == "delay":
                    seconds = inner_step.params.get("seconds", 0)
                    if seconds > 0:
                        await workflow.sleep(timedelta(seconds=seconds))
                    continue

                context_data = await workflow.execute_activity(
                    execute_step,
                    args=[
                        run_id,
                        exe_step.index,
                        inner_step.kind,
                        self._serialize_params(inner_step.params),
                        f"{repr(inner_step)} [iter {i}/{n}]",
                        exe_step.phase.value,
                        context_data,
                        nuon_base_url,
                        nuon_api_token,
                    ],
                    start_to_close_timeout=timedelta(minutes=30),
                    retry_policy=workflow.RetryPolicy(maximum_attempts=1),
                )

            # Sleep between iterations (not after the last one)
            if interval_seconds > 0 and i < n:
                await workflow.sleep(timedelta(seconds=interval_seconds))

        return context_data

    @staticmethod
    def _serialize_params(params: dict[str, Any]) -> dict[str, Any]:
        """Make params JSON-serializable for Temporal."""
        result = {}
        for k, v in params.items():
            if callable(v):
                result[k] = f"<callable:{getattr(v, '__name__', 'lambda')}>"
            elif isinstance(v, list) and v and hasattr(v[0], "kind"):
                # List of Step objects inside repeat() — skip, handled by workflow
                continue
            elif hasattr(v, "__dataclass_fields__"):
                # Dataclass (App, Install, Component) — convert to dict
                from dataclasses import asdict
                result[k] = asdict(v)
            else:
                result[k] = v
        return result

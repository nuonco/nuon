"""API routes — trigger runs, inject faults, cancel runs."""

from __future__ import annotations

import os
from typing import Any

from fastapi import APIRouter, HTTPException
from nanoid import generate as nanoid
from pydantic import BaseModel
from temporalio.client import Client

from cerberus.db.models import CanaryRun
from cerberus.db.session import get_session
from cerberus.dsl.registry import get_canary, list_canaries
from cerberus.temporal.workflows import TASK_QUEUE, CanaryRunWorkflow

router = APIRouter(prefix="/api")


class TriggerRunRequest(BaseModel):
    canary_name: str
    params: dict[str, Any] = {}


class InjectFaultRequest(BaseModel):
    fault_type: str
    config: dict[str, Any] = {}


@router.get("/canaries")
async def api_list_canaries():
    """List all registered canary definitions."""
    return [
        {
            "name": c.name,
            "description": c.description,
            "category": c.category,
            "timeout_seconds": int(c.timeout.total_seconds()),
            "num_installs": len(c.installs),
            "num_steps": len(c.steps),
        }
        for c in list_canaries()
    ]


@router.post("/runs")
async def api_trigger_run(req: TriggerRunRequest):
    """Trigger a new canary run."""
    try:
        defn = get_canary(req.canary_name)
    except KeyError as e:
        raise HTTPException(status_code=404, detail=str(e))

    run_id = nanoid(size=26)

    nuon_base_url = os.environ.get("NUON_API_URL", "http://localhost:8081")
    nuon_api_token = os.environ.get("NUON_API_TOKEN", "")

    if not nuon_api_token:
        raise HTTPException(status_code=500, detail="NUON_API_TOKEN not configured")

    # Create run record in DB
    async with get_session() as db:
        run = CanaryRun(
            id=run_id,
            canary_name=req.canary_name,
            status="pending",
            params=req.params,
        )
        db.add(run)

    # Start Temporal workflow
    temporal_address = os.environ.get("TEMPORAL_ADDRESS", "localhost:7233")
    temporal_namespace = os.environ.get("TEMPORAL_NAMESPACE", "cerberus")

    client = await Client.connect(temporal_address, namespace=temporal_namespace)
    workflow_id = f"cerberus-{run_id}"

    handle = await client.start_workflow(
        CanaryRunWorkflow.run,
        args=[run_id, req.canary_name, req.params, nuon_base_url, nuon_api_token],
        id=workflow_id,
        task_queue=TASK_QUEUE,
        execution_timeout=defn.timeout,
    )

    # Update run with Temporal IDs
    async with get_session() as db:
        run = await db.get(CanaryRun, run_id)
        if run:
            run.temporal_workflow_id = workflow_id
            run.temporal_run_id = handle.result_run_id

    return {"run_id": run_id, "workflow_id": workflow_id}


@router.post("/runs/{run_id}/fault")
async def api_inject_fault(run_id: str, req: InjectFaultRequest):
    """Inject a fault into a running canary via Temporal signal."""
    async with get_session() as db:
        run = await db.get(CanaryRun, run_id)
        if not run:
            raise HTTPException(status_code=404, detail="Run not found")
        if run.status != "running":
            raise HTTPException(status_code=400, detail=f"Run is {run.status}, not running")
        workflow_id = run.temporal_workflow_id

    temporal_address = os.environ.get("TEMPORAL_ADDRESS", "localhost:7233")
    temporal_namespace = os.environ.get("TEMPORAL_NAMESPACE", "cerberus")

    client = await Client.connect(temporal_address, namespace=temporal_namespace)
    handle = client.get_workflow_handle(workflow_id)
    await handle.signal(CanaryRunWorkflow.inject_fault, {"type": req.fault_type, **req.config})

    return {"run_id": run_id, "fault_injected": True}


@router.post("/runs/{run_id}/cancel")
async def api_cancel_run(run_id: str):
    """Cancel a running canary via Temporal signal."""
    async with get_session() as db:
        run = await db.get(CanaryRun, run_id)
        if not run:
            raise HTTPException(status_code=404, detail="Run not found")
        if run.status != "running":
            raise HTTPException(status_code=400, detail=f"Run is {run.status}, not running")
        workflow_id = run.temporal_workflow_id

    temporal_address = os.environ.get("TEMPORAL_ADDRESS", "localhost:7233")
    temporal_namespace = os.environ.get("TEMPORAL_NAMESPACE", "cerberus")

    client = await Client.connect(temporal_address, namespace=temporal_namespace)
    handle = client.get_workflow_handle(workflow_id)
    await handle.signal(CanaryRunWorkflow.cancel_run)

    return {"run_id": run_id, "cancelled": True}

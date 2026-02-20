"""Run detail routes — view a canary run's timeline, steps, and events."""

from __future__ import annotations

from fastapi import APIRouter, Request, HTTPException
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from cerberus.db.models import CanaryRun, RunStep, RunEvent
from cerberus.db.session import get_session

router = APIRouter()


@router.get("/runs/{run_id}")
async def run_detail(request: Request, run_id: str):
    templates = request.app.state.templates

    async with get_session() as db:
        result = await db.execute(
            select(CanaryRun)
            .options(selectinload(CanaryRun.steps), selectinload(CanaryRun.events))
            .where(CanaryRun.id == run_id)
        )
        run = result.scalar_one_or_none()

    if not run:
        raise HTTPException(status_code=404, detail="Run not found")

    return templates.TemplateResponse(
        "run_detail.html",
        {"request": request, "run": run},
    )


@router.get("/runs/{run_id}/timeline")
async def run_timeline_partial(request: Request, run_id: str):
    """HTMX partial — returns just the timeline fragment for live updates."""
    templates = request.app.state.templates

    async with get_session() as db:
        result = await db.execute(
            select(CanaryRun)
            .options(selectinload(CanaryRun.steps), selectinload(CanaryRun.events))
            .where(CanaryRun.id == run_id)
        )
        run = result.scalar_one_or_none()

    if not run:
        raise HTTPException(status_code=404, detail="Run not found")

    return templates.TemplateResponse(
        "partials/run_timeline.html",
        {"request": request, "run": run},
    )


@router.get("/runs/{run_id}/events")
async def run_events_partial(request: Request, run_id: str):
    """HTMX partial — returns the event log fragment."""
    templates = request.app.state.templates

    async with get_session() as db:
        result = await db.execute(
            select(RunEvent)
            .where(RunEvent.run_id == run_id)
            .order_by(RunEvent.created_at.desc())
            .limit(100)
        )
        events = result.scalars().all()

    return templates.TemplateResponse(
        "partials/event_log.html",
        {"request": request, "events": events, "run_id": run_id},
    )

"""Dashboard route — main page showing canary list and recent runs."""

from __future__ import annotations

from fastapi import APIRouter, Request
from sqlalchemy import select, func

from cerberus.db.models import CanaryRun
from cerberus.db.session import get_session
from cerberus.dsl.registry import list_canaries

router = APIRouter()


@router.get("/")
async def dashboard(request: Request):
    templates = request.app.state.templates
    canaries = list_canaries()

    async with get_session() as db:
        # Recent runs
        result = await db.execute(
            select(CanaryRun).order_by(CanaryRun.created_at.desc()).limit(20)
        )
        recent_runs = result.scalars().all()

        # Stats
        total_result = await db.execute(select(func.count(CanaryRun.id)))
        total_runs = total_result.scalar() or 0

        passed_result = await db.execute(
            select(func.count(CanaryRun.id)).where(CanaryRun.status == "passed")
        )
        passed_runs = passed_result.scalar() or 0

        failed_result = await db.execute(
            select(func.count(CanaryRun.id)).where(CanaryRun.status == "failed")
        )
        failed_runs = failed_result.scalar() or 0

        running_result = await db.execute(
            select(func.count(CanaryRun.id)).where(CanaryRun.status == "running")
        )
        running_runs = running_result.scalar() or 0

    return templates.TemplateResponse(
        "dashboard.html",
        {
            "request": request,
            "canaries": canaries,
            "recent_runs": recent_runs,
            "stats": {
                "total": total_runs,
                "passed": passed_runs,
                "failed": failed_runs,
                "running": running_runs,
            },
        },
    )

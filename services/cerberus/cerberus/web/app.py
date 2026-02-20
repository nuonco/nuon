"""FastAPI web application for Cerberus UI."""

from __future__ import annotations

import logging
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

from cerberus.db.session import init_db
from cerberus.dsl.registry import discover_canaries

logger = logging.getLogger(__name__)

TEMPLATES_DIR = os.path.join(os.path.dirname(__file__), "templates")
STATIC_DIR = os.path.join(os.path.dirname(__file__), "static")


@asynccontextmanager
async def lifespan(app: FastAPI):
    discover_canaries()
    await init_db()
    yield


def create_app() -> FastAPI:
    app = FastAPI(title="Cerberus", lifespan=lifespan)
    app.mount("/static", StaticFiles(directory=STATIC_DIR), name="static")

    templates = Jinja2Templates(directory=TEMPLATES_DIR)
    app.state.templates = templates

    from cerberus.web.routes.dashboard import router as dashboard_router
    from cerberus.web.routes.runs import router as runs_router
    from cerberus.web.routes.api import router as api_router

    app.include_router(dashboard_router)
    app.include_router(runs_router)
    app.include_router(api_router)

    return app


app = create_app()

if __name__ == "__main__":
    import uvicorn

    logging.basicConfig(level=logging.INFO)
    uvicorn.run("cerberus.web.app:app", host="0.0.0.0", port=8090, reload=True)

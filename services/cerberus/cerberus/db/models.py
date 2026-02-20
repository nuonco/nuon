"""SQLAlchemy models for Cerberus."""

from __future__ import annotations

from datetime import datetime, timezone

from sqlalchemy import JSON, DateTime, ForeignKey, Index, Integer, String, Text, func
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship


class Base(DeclarativeBase):
    pass


def _utcnow() -> datetime:
    return datetime.now(timezone.utc)


class CanaryRun(Base):
    __tablename__ = "canary_runs"

    id: Mapped[str] = mapped_column(String(30), primary_key=True)
    canary_name: Mapped[str] = mapped_column(String(255), nullable=False, index=True)
    status: Mapped[str] = mapped_column(
        String(20), nullable=False, default="pending", index=True
    )  # pending | running | passed | failed | cancelled | timed_out
    params: Mapped[dict | None] = mapped_column(JSON, nullable=True)

    temporal_workflow_id: Mapped[str | None] = mapped_column(String(255), nullable=True)
    temporal_run_id: Mapped[str | None] = mapped_column(String(255), nullable=True)

    # Nuon resources created during this run
    nuon_org_id: Mapped[str | None] = mapped_column(String(50), nullable=True)
    nuon_app_id: Mapped[str | None] = mapped_column(String(50), nullable=True)
    nuon_install_ids: Mapped[dict | None] = mapped_column(JSON, nullable=True)  # name -> id

    error_message: Mapped[str | None] = mapped_column(Text, nullable=True)
    error_step: Mapped[str | None] = mapped_column(String(255), nullable=True)

    started_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    completed_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, default=_utcnow, server_default=func.now()
    )

    steps: Mapped[list[RunStep]] = relationship(
        back_populates="run", cascade="all, delete-orphan", order_by="RunStep.step_index"
    )
    events: Mapped[list[RunEvent]] = relationship(
        back_populates="run", cascade="all, delete-orphan", order_by="RunEvent.created_at"
    )

    __table_args__ = (
        Index("ix_canary_runs_created_at", "created_at"),
    )


class RunStep(Base):
    __tablename__ = "run_steps"

    id: Mapped[str] = mapped_column(String(30), primary_key=True)
    run_id: Mapped[str] = mapped_column(
        String(30), ForeignKey("canary_runs.id", ondelete="CASCADE"), nullable=False, index=True
    )
    step_index: Mapped[int] = mapped_column(Integer, nullable=False)
    name: Mapped[str] = mapped_column(String(500), nullable=False)
    phase: Mapped[str] = mapped_column(String(20), nullable=False)  # setup | steps | cleanup
    status: Mapped[str] = mapped_column(
        String(20), nullable=False, default="pending"
    )  # pending | running | passed | failed | skipped

    input_data: Mapped[dict | None] = mapped_column(JSON, nullable=True)
    output_data: Mapped[dict | None] = mapped_column(JSON, nullable=True)
    error_message: Mapped[str | None] = mapped_column(Text, nullable=True)
    error_traceback: Mapped[str | None] = mapped_column(Text, nullable=True)

    repeat_iteration: Mapped[int | None] = mapped_column(Integer, nullable=True)
    retry_attempt: Mapped[int] = mapped_column(Integer, nullable=False, default=0)

    started_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    completed_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    duration_ms: Mapped[int | None] = mapped_column(Integer, nullable=True)

    run: Mapped[CanaryRun] = relationship(back_populates="steps")
    events: Mapped[list[RunEvent]] = relationship(
        back_populates="step", cascade="all, delete-orphan"
    )


class RunEvent(Base):
    __tablename__ = "run_events"

    id: Mapped[str] = mapped_column(String(30), primary_key=True)
    run_id: Mapped[str] = mapped_column(
        String(30), ForeignKey("canary_runs.id", ondelete="CASCADE"), nullable=False, index=True
    )
    step_id: Mapped[str | None] = mapped_column(
        String(30), ForeignKey("run_steps.id", ondelete="CASCADE"), nullable=True
    )
    event_type: Mapped[str] = mapped_column(
        String(30), nullable=False
    )  # api_call | api_response | fault_injected | status_change | error | log
    message: Mapped[str] = mapped_column(Text, nullable=False)
    data: Mapped[dict | None] = mapped_column(JSON, nullable=True)
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, default=_utcnow, server_default=func.now()
    )

    run: Mapped[CanaryRun] = relationship(back_populates="events")
    step: Mapped[RunStep | None] = relationship(back_populates="events")

    __table_args__ = (
        Index("ix_run_events_created_at", "created_at"),
    )

"""Temporal worker for Cerberus."""

from __future__ import annotations

import asyncio
import logging
import os

from temporalio.client import Client
from temporalio.worker import Worker

from cerberus.dsl.registry import discover_canaries
from cerberus.temporal.activities import execute_step, save_run_context, update_run_status
from cerberus.temporal.workflows import TASK_QUEUE, CanaryRunWorkflow

logger = logging.getLogger(__name__)


async def run_worker() -> None:
    """Start the Cerberus Temporal worker."""
    # Discover all canary definitions
    discover_canaries()

    temporal_address = os.environ.get("TEMPORAL_ADDRESS", "localhost:7233")
    temporal_namespace = os.environ.get("TEMPORAL_NAMESPACE", "cerberus")

    logger.info(f"Connecting to Temporal at {temporal_address} (namespace: {temporal_namespace})")
    client = await Client.connect(temporal_address, namespace=temporal_namespace)

    worker = Worker(
        client,
        task_queue=TASK_QUEUE,
        workflows=[CanaryRunWorkflow],
        activities=[execute_step, update_run_status, save_run_context],
    )

    logger.info(f"Starting Cerberus worker on task queue '{TASK_QUEUE}'")
    await worker.run()


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(run_worker())

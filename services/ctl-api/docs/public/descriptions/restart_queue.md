# Restart Queue Workflow

This admin endpoint restarts a queue's Temporal workflow by sending a restart update signal.

## What This Does

When you restart a queue:
1. The system sends a restart update to the queue's Temporal workflow
2. If the workflow is running, it receives the restart signal
3. If the workflow is not running, Temporal starts it and then sends the restart signal
4. The queue workflow sets its internal `restarted` flag to true

## When to Use This

Use this endpoint to:
- Recover a queue workflow that has stopped processing signals
- Reset a queue's state after configuration changes
- Force a queue workflow to restart its processing loop

## Authentication

This endpoint requires admin authentication via the `X-Nuon-Admin-Email` header.

## Side Effects

- The queue workflow will receive a restart update
- If the workflow wasn't running, it will be started
- The workflow's internal state will be updated to reflect the restart

## Example

```bash
curl -X POST \
  https://api.nuon.co/v1/queues/{queue_id}/admin-restart \
  -H "X-Nuon-Admin-Email: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{}'
```

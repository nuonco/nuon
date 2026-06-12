# 001 — Runner Health Check

## Problem

Runner health is monitored at the process level (`process_healthcheck`, every 5-15 min per process) but there is no
periodic runner-level health check that aggregates process state into runner status. The existing `offline_check` signal
uses a 24-hour stale threshold and is only triggered manually via admin endpoint. This means a runner can have no active
processes but still show as "active" indefinitely.

## Requirements

1. A `runner_healthcheck` signal that fires every 15 minutes via a cron emitter on the `runner-signals` queue.
2. For **org runners**: check for an active `org` process. If missing, mark runner offline.
3. For **install runners**: check for an active `install` process (primary). If missing, mark runner offline.
   Additionally check for an active `mng` process — if missing, set `missing_mng_process` metadata warning on the
   runner's v2 status.
4. Skip runners in transitional states (provisioning, deprovisioning, reprovisioning, deprovisioned, pending).
5. Emit `runner.health_check` counter metric with tags for runner type, org, missing processes, and result
   (healthy/unhealthy/skipped).
6. Emit a Datadog event when a runner transitions from active to offline, including owner context (org name, app name,
   install name, creator email for install runners).

## Design Decisions

- **Primary process per runner type**: Org runners use `org` process, install runners use `install` process. The `mng`
  process for install runners is a secondary check that produces a metadata warning but does not determine
  active/offline status.
- **Emitter placement**: Cron emitter is created alongside the `runner-signals` queue — in `CreateOrgRunnerGroup` for
  org runners and `EnsureRunnerSignalsQueue` for install runners (with idempotent check).
- **MaxInFlightAge**: 10 minutes. Prevents stale signals from accumulating if the queue backs up.
- **Status update**: Only calls `UpdateStatus` (v1) when the status actually changes. Always calls
  `UpdateRunnerStatusV2` (v2) to keep metadata current.
- **Datadog event enrichment**: For install runners, fetches the full install (with App, Org, CreatedBy preloads) to
  include app name, install name, and creator email in the event body.

## Key Files

### New
- `services/ctl-api/internal/app/runners/signals/runnerhealthcheck/signal.go` — signal logic
- `services/ctl-api/internal/app/runners/signals/runnerhealthcheck/init.go` — catalog registration

### Modified
- `services/ctl-api/internal/app/runners/helpers/create_runner_group.go` — emitter for org runners
- `services/ctl-api/internal/app/runners/helpers/ensure_runner_queues.go` — emitter for install runners
- `services/ctl-api/internal/pkg/workflows/status/activities/update.go` — Metadata field on UpdateRunnerStatusV2Request
- `services/ctl-api/internal/pkg/queue/catalog/allsignals/allsignals.go` — signal registration

## Metrics

| Metric | Type | Tags |
|--------|------|------|
| `runner.health_check` | Counter | `runner_id`, `runner_type`, `runner_status`, `org_id`, `org_name`, `install_id`, `result` (healthy/unhealthy/skipped), `missing_org_process`, `missing_install_process`, `missing_mng_process` |

## Datadog Event (active -> offline)

| Field | Value |
|-------|-------|
| Title | `Runner {display_name} went offline` |
| Alert Type | Error |
| Tags | `runner_id`, `runner_type`, `org_id`, `org_name`, `install_id`, `install_name`, `app_id`, `app_name`, `created_by` |
| Aggregation Key | `runner-health-check` |

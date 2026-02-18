# Runner Concepts

Core concepts for runner groups, leader election, tainting, and job routing.

## Runner Groups

Runner groups are polymorphic — they can be owned by an `Org` or an `Install`:

```
RunnerGroup
├── OwnerID    (org ID or install ID)
├── OwnerType  ("org" or "install")
├── LeaderRunnerID  (*string, nullable)
└── Runners[]
```

Each group contains one or more `Runner` records. The `ActiveRunner()` helper on `RunnerGroup` returns the leader if
set, otherwise falls back to the first runner.

A group can have multiple runners with different platforms (e.g. `aws-eks`, `local`):

- **Only one runner is the leader** — all workload jobs route to it.
- **Non-leader runners** still have running event loops and health checks, but `RetargetJobToLeader` redirects
  workload jobs away from them.
- **Platform diversity**: a group can mix cloud and local runners. Tainting controls which platforms are eligible
  for leadership.
- **Runner creation**: `AdminCreateRunnerInGroup` is idempotent per platform — if a runner with the requested
  platform already exists, it returns the existing one with a fresh token.
- **Local runners are immediately active** (`RunnerStatusActive`) since they skip cloud provisioning. Cloud runners
  start as `pending` and transition to `active` after provisioning.

## Leader Election

**Location**: `services/ctl-api/internal/app/runners/helpers/elect_leader.go`

The control plane decides the leader — runners themselves have no awareness of the election. Only one runner per group
is the leader at any time.

`ElectLeader(ctx, groupID)` runs a serialized database transaction:

1. **Lock the runner group** with `SELECT ... FOR UPDATE` to prevent concurrent elections.
2. **Query for the best candidate**: the oldest (`created_at ASC`) active, untainted runner in the group.
   ```sql
   WHERE runner_group_id = ? AND status = 'active' AND tainted = false AND deleted_at = 0
   ORDER BY created_at ASC
   LIMIT 1
   ```
3. **Update `leader_runner_id`**:
   - If a candidate exists → set it as leader.
   - If no candidates exist → clear the leader (`leader_runner_id = NULL`).
4. **Commit the transaction.**
5. **Reschedule jobs** (post-commit) if the leader changed — see [Job Rescheduling](#job-rescheduling-on-leader-change).

### Election Triggers

| Trigger | Location | When |
|---------|----------|------|
| **Runner creation** | `AdminCreateRunnerInGroup` | After a new runner is created in a group |
| **Health check status change** | `cron_health_check.go` | When the current leader becomes unhealthy, or a runner becomes active and the group has no leader |

### Health Check Integration

The cron health check (`HealthCheck` workflow, runs every minute per runner) triggers election when:

- The **current leader becomes unhealthy** (status != `active`) — elect a new leader to take over.
- A **runner becomes active** and the group has **no leader** — elect it as leader.

```go
// Leader became unhealthy
if newStatus != RunnerStatusActive && group.LeaderRunnerID == runner.ID {
    ElectLeader(ctx, runner.RunnerGroupID)
}
// Runner became active, no leader set
if newStatus == RunnerStatusActive && group.LeaderRunnerID == nil {
    ElectLeader(ctx, runner.RunnerGroupID)
}
```

## Tainting

A runner can be **tainted** (`tainted = true`), which permanently excludes it from leader election until untainted.
Tainting does not stop a runner from operating — it only affects the election query.

| Endpoint | Auth | Description |
|----------|------|-------------|
| `POST /v1/runners/{runner_id}/taint` | API Key + Org | Taint a runner (user-facing) |
| `POST /v1/runners/{runner_id}/untaint` | API Key + Org | Untaint a runner (user-facing) |
| `POST /v1/runners/{runner_id}/taint` (admin) | Admin email | Taint a runner (admin) |
| `POST /v1/runners/{runner_id}/untaint` (admin) | Admin email | Untaint a runner (admin) |

**Location**: `services/ctl-api/internal/app/runners/service/taint_runner.go`, `admin_taint_runner.go`

## Job Rescheduling on Leader Change

When `ElectLeader` commits a new leader that differs from the old leader, queued jobs are moved from the old leader
to the new one.

**Location**: `services/ctl-api/internal/app/runners/helpers/elect_leader.go` (`rescheduleJobsToLeader`)

```
1. Query all jobs on the old leader with status = "queued".
2. For each job:
   a. UPDATE runner_id to the new leader.
   b. Signal the new leader's event loop: OperationProcessJob with the job ID.
```

This runs **outside** the election transaction (post-commit) to avoid holding the lock during signaling. Signals use
`evClient.Send()` which is fire-and-forget — if the event loop isn't running yet, the signal is dropped (but the job
remains reassigned in the database).

### RetargetJobToLeader (Backstop)

Even without rescheduling, `ProcessJob` has a built-in backstop via the `RetargetJobToLeader` activity.

**Location**: `services/ctl-api/internal/app/runners/worker/activities/retarget_job_to_leader.go`

Before processing, `ProcessJob` calls `RetargetJobToLeader` which:

1. Looks up the runner's group and its current `leader_runner_id`.
2. If **no leader** → returns `NoLeader: true`, job is marked `not_attempted`.
3. If **this runner is the leader** → proceeds normally.
4. If **another runner is the leader** → updates `runner_id` on the job to the leader, returns `Retargeted: true`.
   The workflow exits; the job will be picked up by the leader's event loop.

## Job Lifecycle with Leader Election

```
Job Created (status: queued, runner_id: leader)
    │
    ▼
Event loop receives OperationProcessJob signal
    │
    ▼
ProcessJob workflow starts
    │
    ├─ Runner unhealthy? → mark not_attempted, exit
    │
    ├─ RetargetJobToLeader
    │   ├─ No leader? → mark not_attempted, exit
    │   ├─ Already leader? → continue
    │   └─ Not leader? → retarget job, exit
    │
    ▼
Flush orphaned jobs, check timeouts, execute job
```

## Key Files Reference

| File | Description |
|------|-------------|
| `services/ctl-api/internal/app/runners/helpers/elect_leader.go` | Election algorithm and job rescheduling |
| `services/ctl-api/internal/app/runners/helpers/helpers.go` | Helpers struct with dependencies |
| `services/ctl-api/internal/app/runners/service/admin_create_runner_in_group.go` | Runner creation + election trigger |
| `services/ctl-api/internal/app/runners/service/taint_runner.go` | Taint/untaint endpoints |
| `services/ctl-api/internal/app/runners/service/admin_taint_runner.go` | Admin taint/untaint endpoints |
| `services/ctl-api/internal/app/runners/worker/process_job.go` | Job processing with leader check |
| `services/ctl-api/internal/app/runners/worker/activities/retarget_job_to_leader.go` | Job retargeting activity |
| `services/ctl-api/internal/app/runners/worker/cron_health_check.go` | Health check with election trigger |
| `services/ctl-api/internal/app/runners/signals/signals.go` | Signal types (OperationProcessJob, etc.) |
| `services/ctl-api/internal/app/runner_group.go` | RunnerGroup model with LeaderRunnerID |
| `services/ctl-api/internal/app/runner.go` | Runner model with Tainted field |
| `services/ctl-api/internal/app/runner_job.go` | RunnerJob model and statuses |
| `services/ctl-api/internal/pkg/eventloop/send.go` | Event loop signal delivery |

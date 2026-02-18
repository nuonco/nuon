# Runner Upgrades

Runners within a RunnerGroup need to be upgraded (new container image, new version) without dropping in-flight jobs.
This doc defines both a **manual upgrade flow** (user controls each step) and an **automatic rolling upgrade**
(Temporal workflow orchestrates sequentially). Both flows use the same underlying primitives.

See [concepts.md](concepts.md) for leader election, tainting, and job rescheduling.

## Prerequisites

These changes to existing code are required before either upgrade flow works correctly.

### 1. Management/operations jobs bypass leader election

Today, non-leader runners can never pick up their own update/shutdown/health-check jobs because both enforcement points
block them. Per-runner operations must be exempt from leader checks.

**`GetRunnerJobs` (get_runner_jobs.go)** — only enforce leader for workload job groups:

```go
isWorkloadGroup := grp != app.RunnerJobGroupOperations &&
    grp != app.RunnerJobGroupManagement &&
    grp != app.RunnerJobGroupHealthChecks
if isWorkloadGroup {
    if runner.RunnerGroup.LeaderRunnerID != nil && *runner.RunnerGroup.LeaderRunnerID != runnerID {
        return []*app.RunnerJob{}, nil
    }
}
```

**`ProcessJob` (process_job.go)** — skip `RetargetJobToLeader` for per-runner job types:

```go
if runnerJob.Group != app.RunnerJobGroupOperations &&
    runnerJob.Group != app.RunnerJobGroupManagement &&
    runnerJob.Group != app.RunnerJobGroupHealthChecks {
    // existing leader election check (RetargetJobToLeader)...
}
```

Note: the `ProcessJob` change requires fetching the job _before_ the leader check (move `AwaitGetJob` up), or passing
the job group through the signal. The simpler approach is to move the leader check after `AwaitGetJob` and use the
job's group to decide.

### 2. Taint triggers immediate leader election

Currently `TaintRunner` and `UntaintRunner` just flip a boolean. If the tainted runner is the current leader, it stays
leader until the next health check cycle (~1 minute). Both endpoints must call `ElectLeader` when the change could
affect leadership.

**`TaintRunner`** — after setting `tainted = true`:

```go
// If this runner was the leader, elect a new one immediately.
if err := s.db.WithContext(ctx).Preload("RunnerGroup").First(&runner, "id = ?", runnerID).Error; err == nil {
    if runner.RunnerGroup.LeaderRunnerID != nil && *runner.RunnerGroup.LeaderRunnerID == runnerID {
        s.runnersHelpers.ElectLeader(ctx, runner.RunnerGroupID)
    }
}
```

**`UntaintRunner`** — after setting `tainted = false`:

```go
// If the group has no leader, this runner might be eligible.
if err := s.db.WithContext(ctx).Preload("RunnerGroup").First(&runner, "id = ?", runnerID).Error; err == nil {
    if runner.RunnerGroup.LeaderRunnerID == nil {
        s.runnersHelpers.ElectLeader(ctx, runner.RunnerGroupID)
    }
}
```

### 3. Fix `UpdateVersion` hardcoding `Runners[0]`

`update_version.go:38` always targets `runner.Org.RunnerGroup.Runners[0]`. When re-enabled, this must target the
specific runner that reported the version mismatch (the runner whose health check triggered the workflow). The runner ID
is already available via `sreq.ID`.

### 4. Per-runner version override (new model field)

`ContainerImageTag` lives on `RunnerGroupSettings` (shared across all runners in a group). To upgrade runners
individually, we need a per-runner version field.

**Add to `Runner` model:**

```go
// TargetVersion overrides RunnerGroupSettings.ContainerImageTag for this specific runner.
// When set, Helm upgrade / MngUpdate uses this value. When empty, falls back to group settings.
TargetVersion string `json:"target_version,omitzero" gorm:"default:null" temporaljson:"target_version,omitzero,omitempty"`
```

**Resolution order** (in Helm upgrade / reprovision activities):

1. `Runner.TargetVersion` if non-empty
2. `RunnerGroupSettings.ContainerImageTag` (group default)

This allows upgrading Runner B to v1.2.3 while A and C stay on v1.1.0.

## Layer 1: Manual Upgrade Primitives

The manual flow gives the user full control over each step. All primitives already exist or need minor additions.

### User Flow

```
Group has: Runner A (leader, oldest), Runner B, Runner C

# Upgrade a non-leader first to test
1. nuon runners taint B                      # B excluded from leader election
2. nuon runners upgrade B --version v1.2.3   # reprovision just B with new version
3.   ... B comes back active, user tests ...
4.   ... looks good ...
5. nuon runners untaint B                    # B eligible for leadership again

# Upgrade next non-leader
6. nuon runners taint C
7. nuon runners upgrade C --version v1.2.3
8.   ... C comes back, tests pass ...
9. nuon runners untaint C

# Upgrade the leader last
10. nuon runners taint A                     # leadership moves to B or C immediately
11. nuon runners upgrade A --version v1.2.3
12.   ... A comes back active ...
13. nuon runners untaint A                   # A re-eligible for leadership
```

The user can stop at any point. If step 3 reveals a problem, they reprovision B back to the old version or investigate.
The leader never changes unless the user explicitly taints it.

### New API Endpoint

```
POST /v1/runners/:runner_id/upgrade
Body: { "target_version": "v1.2.3" }
Response: 202 { "operation_id": "..." }
```

Implementation:

1. Set `Runner.TargetVersion = req.TargetVersion`
2. Determine upgrade method based on runner platform:
   - K8s runners → send `OperationReprovision` signal
   - VM runners → send `OperationMngUpdate` signal
3. Return operation ID for tracking

### Existing Endpoints (No Changes Needed)

| Action | Endpoint | Notes |
|--------|----------|-------|
| Taint runner | `POST /v1/runners/:id/taint` | Add ElectLeader call (prereq 2) |
| Untaint runner | `POST /v1/runners/:id/untaint` | Add ElectLeader call (prereq 2) |
| View leader | `GET /v1/runner-groups/:id/leader` | Already exists |
| Set leader manually | `PUT /v1/runner-groups/:id/leader` | Already exists |
| Check runner status | `GET /v1/runners/:id` | Heartbeat already reports version |

### CLI Commands

```bash
nuon runners taint <runner-id>
nuon runners untaint <runner-id>
nuon runners upgrade <runner-id> --version <tag>
nuon runners get <runner-id>                # shows version, tainted, leader status
nuon runner-groups get-leader <group-id>
```

## Layer 2: Automatic Rolling Upgrade

A Temporal workflow that calls the same primitives from Layer 1 in sequence. Users who trust the process use this
instead of manual steps.

### API Endpoint

```
POST /v1/runner-groups/:runner_group_id/rolling-upgrade
Body: { "target_version": "v1.2.3" }
Response: 202 { "operation_id": "..." }
```

### Workflow: `RollingUpgrade`

```
RollingUpgrade(groupID, targetVersion):
  1. Load group + runners
  2. Validate: at least 1 active runner, no upgrade already in progress
  3. Sort runners: non-leaders first (by created_at DESC), current leader last
  4. For each runner:
     a. Taint runner
     b. ElectLeader (leadership moves away if this was the leader)
     c. Drain: poll until runner has no in-progress jobs (with timeout from OverallTimeout)
     d. Set Runner.TargetVersion = targetVersion
     e. Signal Reprovision (K8s) or MngUpdate (VM)
     f. Poll until runner status = active AND heartbeat reports targetVersion (with timeout)
     g. Untaint runner
     h. ElectLeader (runner re-eligible)
     i. On timeout/failure at any step: stop, leave runner tainted, report partial failure
  5. Update RunnerGroupSettings.ContainerImageTag = targetVersion (group default now matches)
  6. Clear Runner.TargetVersion on all runners (no longer needed, group default matches)
```

### Failure Handling

- If a runner fails to come back after upgrade, the workflow **stops** and leaves that runner tainted.
- Already-upgraded runners continue serving (they have the new version).
- Not-yet-upgraded runners continue serving (they have the old version).
- The user can investigate, fix the problem, then either:
  - Resume: `POST /v1/runner-groups/:id/rolling-upgrade` (workflow picks up remaining runners)
  - Rollback: manually upgrade the failed runner back to old version and untaint it

### Single-Runner Groups

The workflow degenerates gracefully: taint → upgrade → untaint. There is a brief downtime window during reprovision,
identical to today's behavior. No special-casing needed.

### Cancellation

The user can cancel a rolling upgrade at any time. The workflow stops after the current runner finishes (or times out).
Already-upgraded runners keep their new version. The current runner is left tainted if mid-upgrade.

```
DELETE /v1/runner-groups/:runner_group_id/rolling-upgrade
```

### Visibility

```
GET /v1/runner-groups/:runner_group_id/upgrade-status
Response: {
  "status": "in-progress",    // idle | in-progress | failed | completed
  "target_version": "v1.2.3",
  "total_runners": 3,
  "upgraded_runners": 1,
  "current_runner_id": "rnr...",
  "current_phase": "draining", // tainting | draining | upgrading | waiting-active | untainting
  "failed_runner_id": null
}
```

## Install Runners

`Reprovision` currently rejects install runners (reprovision.go:72). VM-based install runners can use the `MngUpdate`
path. K8s install runners need `Reprovision` to be extended — this is a separate effort and out of scope for this doc.

## Codebase Locations

| Component | File |
|-----------|------|
| Runner model (add TargetVersion) | `services/ctl-api/internal/app/runner.go` |
| Manual upgrade endpoint | `services/ctl-api/internal/app/runners/service/upgrade_runner.go` |
| Rolling upgrade endpoint | `services/ctl-api/internal/app/runners/service/rolling_upgrade.go` |
| Rolling upgrade workflow | `services/ctl-api/internal/app/runners/worker/rolling_upgrade.go` |
| Signal type | `services/ctl-api/internal/app/runners/signals/signals.go` |
| Event loop registration | `services/ctl-api/internal/app/runners/worker/event_loop_workflow.go` |
| Bypass leader for ops jobs | `services/ctl-api/internal/app/runners/service/get_runner_jobs.go` |
| Bypass leader in ProcessJob | `services/ctl-api/internal/app/runners/worker/process_job.go` |
| Taint triggers election | `services/ctl-api/internal/app/runners/service/taint_runner.go` |
| CLI commands | `bins/cli/internal/services/installs/runners.go` |

## Implementation Order

1. **Prerequisites** (1-4) — unblock per-runner operations
2. **Layer 1: Manual primitives** — `POST /v1/runners/:id/upgrade` endpoint + CLI
3. **Layer 2: Rolling upgrade workflow** — orchestration on top of Layer 1
4. **Dashboard UI** — upgrade controls on runner group detail page (future)

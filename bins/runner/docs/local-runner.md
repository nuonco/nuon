# Local Runner

How the local development runner integrates with leader election and tainting to take over job processing from cloud
runners during development.

See [concepts.md](concepts.md) for the underlying election algorithm and tainting mechanics.

## Self-Registration & Takeover

When a developer starts a local runner (`bins/runner` in dev mode), the runner self-registers and taints all cloud
runners in the same group so it wins the election.

**Location**: `bins/runner/internal/pkg/dev/runner.go`

```
1. List all runners for the watched runner type (org or install).
2. If a local runner already exists and all cloud runners are tainted → skip (idempotent).
3. Call AdminCreateRunnerInGroup(groupID, "local") to find-or-create a local runner.
4. For every other runner in the group that is not local and not already tainted → call TaintRunner(runnerID).
5. AdminCreateRunnerInGroup triggers ElectLeader → local runner becomes leader
   (it's the only untainted active runner).
```

Local runners are immediately active (`RunnerStatusActive`) since they skip cloud provisioning. This ensures the local
runner immediately takes over job processing from cloud runners during development.

## Restoring Cloud Leadership

When the local runner is stopped, cloud runners can be untainted (manually or via automation) to restore cloud-based
leadership:

```bash
nuon runners untaint <cloud-runner-id>
```

Untainting triggers `ElectLeader` if the group has no current leader, so the cloud runner resumes leadership
automatically.

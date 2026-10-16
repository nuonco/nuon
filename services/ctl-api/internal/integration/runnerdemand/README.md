# On-demand runner integration tests

This FX suite exercises the control-plane boundary between install workflows and
runner jobs. It uses real PostgreSQL, ClickHouse, Temporal workers, queue handlers,
resident workflow updates, and runner HTTP handlers. Fixtures use `testseed` and
normal GORM creation with foreign keys and hooks enabled.

## Real install runner

`TestInstallRunnerDemand` builds `./bins/runner` once and launches a real
`runner install` subprocess for each case. No management process, container,
cloud instance, or cloud credentials are needed. The subprocess registers its
process, sends heartbeats, polls jobs, fetches the prepared action plan, executes
a local shell script, and reports outputs/status through the production APIs.
Runner requests use real service-account tokens, authentication middleware, and
runner org authorization. An unauthenticated request is checked to return 401.

Both the successful script and an exit-23 failure check that:

- Before startup, the job is available without any process or execution, the
  workflow is waiting for the runner, and its Temporal handlers have closed.
- Pickup marks the original workflow, step, and action run in-progress.
- Actual script stdout reaches ClickHouse and the public log-read endpoint while
  execution is blocked on a test-owned release file, before the script finishes.
- Releasing the script resumes the same workflow/step with the correct outcome
  and exactly one job/execution. Success stores parsed outputs without requiring
  a result record; failure preserves the runner's error metadata.
- Production process initialization makes the install process active, and
  SIGTERM shutdown reports it shut down. No management process is created.

Each subprocess has isolated HOME, TMPDIR, registry storage, and an ephemeral
registry port. Its environment is allowlisted rather than inheriting cloud
credentials. Cleanup stops its process group, with a bounded kill fallback, and
prints captured runner output on failure.

## Controlled runner callbacks

`TestRunnerDemand` retains an HTTP-client runner for deterministic cancellation,
duplicate-callback, deadline, and completion/parking race scenarios. No runner
process or heartbeat is seeded in these cases, so pickup cannot depend on either.

The fixture pre-seeds job plans and uses a test signal that runs successive
phases through the shared job sequence driver. These controlled callbacks use the
existing test router's seeded account/org context rather than runner tokens.
Neither test exercises cloud plan preparation, Helm/Terraform, image pulls,
or cloud instance provisioning.
The real-runner cases verify log ingestion and public reads, not browser streaming
transport.

The controlled cases cover:

- Offline wait parks the flow, group, step, and domain Temporal handlers without
  closing the log stream or starting an execution monitor.
- Pickup makes the existing step and workflow in-progress; two successive jobs
  resume the same workflow/step without cloning jobs or redoing finished phases.
- Duplicate claims conflict; duplicate results return the original result.
- Completion while the original handler is unwinding still resumes successfully.
  A test-only signal barrier holds suspension until the runner result is stored,
  then releases it, so this interleaving does not depend on a timed sleep.
- A parallel runner step finalizes while its sibling awaits manual retry, without
  waiting for the resident idle timer. Downstream work remains pending until the
  retry succeeds, and successful siblings are not re-executed.
- Workflow cancellation before/after pickup rejects claims and late results or
  status changes. Completed deploys remain active, including a dependency owned
  by the same unfinished signal; a finished job whose domain finalization is
  still pending is cancelled with the workflow.
- Cancelling a queued job wakes the parked workflow and fails its step.
- Offline time can exceed the execution budget; that budget starts at claim and
  an expired execution rejects late results.
- Plan and composite-plan fetches refresh repository-scoped GitHub credentials
  without changing saved instructions. Revoked repository access fails closed,
  even if an unrestricted installation token would still be available.

Late output data may still be retained after cancellation; outputs alone do not
change job status or resume a cancelled workflow.

## Production finalization and owner cancellation

`TestRunnerDemandActionResume` adds a real install task-queue worker and executes
the production adhoc action signal from a seeded prepared-job checkpoint. Runner
HTTP callbacks finish the job, and the production signal applies the config,
closes its log stream, and checkpoints finalization. Repeated client resumes and
a direct update-with-start leave the completed action, step, config timestamps,
logs, jobs, and executions unchanged. Cloud plan preparation is still outside
this fixture's scope.

The suite also executes the production deploy-apply signal through the shared
job sequence driver. Prepared checkpoints park without closing logs or applying
the deploy; completion applies the config and regenerates state. An older
`apply.finalized` checkpoint still runs those outer deploy updates. Both cases
retain additional checkpoint fields and reuse the original job on resume.

The same FX suite exercises legacy and composite-status cancellation activities
for deploys, sandbox runs, and action runs. Completed owners are preserved even
when only one status representation has caught up. Cancelling unfinished owners
updates both representations, preserves metadata, and is idempotent. Late
completion cannot resurrect cancelled owners or change their parent projections.

## Run locally

Start the normal local PostgreSQL, ClickHouse, Temporal, and SeaweedFS
dependencies, then run from `services/ctl-api`:

```bash
source tests/integration.env
export DB_NAME=runner_demand_fx_test
export CLICKHOUSE_DB_NAME=runner_demand_fx_test
export GITHUB_APP_KEY="$(openssl genrsa 2048 2>/dev/null)"
export AWS_ENDPOINT_URL_S3=http://127.0.0.1:8333
export AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_EC2_METADATA_DISABLED=true
export BLOB_STORAGE_BUCKET=runner-demand-fx-test BLOB_STORAGE_REGION=us-east-1
curl --fail --request PUT "$AWS_ENDPOINT_URL_S3/$BLOB_STORAGE_BUCKET"
go run ./cmd/nuontest ./internal/integration/runnerdemand -count=1 -v -timeout=10m
```

**`nuontest` drops and recreates the named databases. Use disposable database
names, never development or production databases.** After setup, rerun just the
tests with the same environment using `go test ./internal/integration/runnerdemand`.
Without `INTEGRATION=true` the suite skips.

Each suite run creates and deletes a unique Temporal namespace, stops its FX app
and workers, closes its HTTP servers/client, and uses a temporary blob cache. The
test-scoped Temporal client rejects workflow starts outside that namespace so the
enqueuer cannot start its shared `general`-namespace sweeper. Normal channel-based
dispatch and runner completion/resume signals are real; sweeper recovery is outside
this suite's scope. A test-only GORM create callback routes process queues from the
hardcoded production `runners` namespace into the suite's namespace before normal
creation hooks run. The worker registers production process lifecycle signals and
emitter workflows. No production namespace is used and no foreign keys or hooks
are bypassed. Databases remain available for inspection until explicitly dropped
or recreated by `nuontest`.

The Nuon `ctl-api-integration.yml` flow lane includes this package and runs
packages serially against its migrated database.

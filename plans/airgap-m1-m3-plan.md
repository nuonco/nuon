# Airgap M1–M3 execution plan

Scope agreed with Harsh (2026-08-06): reach M3 fast; local runner only for
M1/M2; defer secrets/inputs resolution and heartbeats; M3 = no phone-home +
runner binary & image packaged in the bundle + real EC2 test in sandbox-ht.

## M1 — plan artifact ships in the bundle

Deliverables:

1. Bundle format: new document layer `plan-envelope.json`
   (`application/vnd.nuon.airgap.plan.v1+json`) alongside provenance and the
   qualification report. `bundle.Open` exposes it.
2. Publish signal renders the envelope at bundle-create:
   - reuses `pkg/runner/airgap.Envelope` (single source of truth with the
     runner);
   - reference install = newest non-deleted install of the bundle's pinned
     app config that has plan-bearing runner jobs (sandbox/deploy/sync);
   - dedup: latest owner (install_sandbox_runs / install_deploys per
     component) wins; linear depends_on chain; apply-plan steps chain via
     plan_from_step to the preceding create-*-plan of the same type;
   - `force_default_cloud_auth: true` (customer env always differs);
   - publish fails with a clear error when no qualifying install exists.
3. `nuon-bundle inspect` prints PLAN + INPUTS sections; `--extract-plan`
   writes the envelope JSON for the offline runner.

Provider-mirror gap from M0: no new code — vendoring already exists behind
the `terraform-provider-mirror` org feature flag
(`services/ctl-api/internal/app/apps/worker/plan/plan_sandbox_build.go`).
E2E enables the flag on the M0 org, rebuilds, and verifies the runner logs
mirror usage instead of `no provider mirror in artifact`.

Demo: create bundle via local ctl-api → download → `nuon-bundle inspect`
shows the plan DAG the runner will execute → `--extract-plan` → local
`runner airgap` replay succeeds with ctl-api unreachable.

## M2 — post-trip results and job outputs (local runner only)

State-store contract (S3-shaped, disk-backed for local test):

- `status.json`: run-level status, per-step results, timestamps.
- `steps/<id>/`: job execution results + outputs (tf outputs etc.).
- Export: after a run, results are collated so a vendor/customer can pull
  them out of the state bucket ("post-trip report").

Explicitly deferred: secrets/input resolution, heartbeats, WaitCondition,
instance-replacement resume demo.

## M3 — no phone-home + packaged runner + EC2

- Bundle packages the runner binary (linux/amd64) and runner container image
  as bundle members.
- Runner starts from bundle + state bucket only: no token, no ctl-api.
- Airgap CF template variant (no phone-home Lambda) or minimal UserData
  bootstrap for the EC2 test.
- Real test: push bundle to S3 in sandbox-ht (504178855485, us-east-1),
  spin up EC2, execute, verify results in state bucket + CloudWatch.
  Leave infra running.

## Constraints

- AWS: only sandbox-ht.NuonAdmin / 504178855485 / us-east-1. Never stage.
- Do not destroy the M0 CF stack or EKS cluster.
- Keep `runner_group_settings.local_awsiam_role_arn = ''` for the M0 runner
  group in the local DB.

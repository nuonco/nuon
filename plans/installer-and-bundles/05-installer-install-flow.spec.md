# 05 — Installer install & update flow

> Owns: the end-to-end install and update flow driven by the installer.
> Depends on: **04** (scaffolding, store, installer-api), **C1** (verify), **C2**
> (jobs, status, approvals). Same binary as 04.

## Goal

Turn a connected installer into something that actually installs and updates an
app: fetch the install's workflows, pull and verify the bundle, show the customer
what will change, get approval, execute, and push results back.

This is the flow that replaces the "stack" semantic — instead of the customer
running a CloudFormation template that provisions a Nuon runner, the installer
executes the work directly.

## Scope

**In**: bundle acquisition and verification, workflow fetch, plan/approve/execute,
result and health push, update (re-install to a newer bundle), resumability.

**Out**: scaffolding and bucket setup (04), the bundle format (01), the
control-plane models and API (03).

## Flow

```
installer install                        installer update
      │                                        │
      ├─ fetch workflows (C2) ────────────────┤   Workflow.ExecutionType == external
      ├─ resolve bundle for the workflow ─────┤
      ├─ download via grant (C2) ─────────────┤
      ├─ VERIFY (C1, fails closed) ───────────┤   transport checksum + blobs + signature
      ├─ push InstallBundle=verified (C2) ────┤
      ├─ claim job (C2, executor=external) ───┤
      ├─ execute plan locally ────────────────┤
      ├─ push plan for approval (C2) ─────────┤
      ├─ customer approves in the web UI ─────┤
      ├─ claim + execute apply ───────────────┤
      ├─ push results + outputs (C2) ─────────┤
      └─ push component health (C2) ──────────┘
```

Steps are driven by the control plane's workflow definition; the installer is a
worker that claims `external` jobs. It does **not** construct plans — everything
executable comes from the bundle.

## 1. Bundle acquisition and verification

1. `GET /v1/installer/bundles` → available bundles; pick the one bound to the
   workflow (or `--bundle-id`).
2. `POST .../download-grants` → `{url, size, transport_checksum, manifest_digest,
   supports_range}`.
3. Download with resume support (validate `Content-Range` on resume).
4. **Verify before touching anything**, via `bundle.VerifyBundle` with a fully
   populated `Expectation`:
   - transport SHA-256 vs. the grant's `transport_checksum`
   - every reachable blob digest (`VerifyBlobs`)
   - the ed25519 signature vs. this install's public key
5. On any failure: delete the download, push `InstallBundle = failed` with the
   reason, exit non-zero. **Never** proceed with a partially verified bundle.
6. On success: push `InstallBundle = verified` with `verified_at`.

Per C1/S1 and S2, a check that was skipped (e.g. no signature available yet) must be
*reported as skipped*, not counted as passed. This is the defect the air-gap
prototype shipped and it is the whole point of the signing work.

## 2. Executing jobs

The installer is a job worker:

```
GET  /v1/installer/jobs?status=available   → claimable jobs (Executor == external)
POST /v1/installer/jobs/{id}/claim         → idempotent per agent
GET  /v1/installer/jobs/{id}/plan          → composite plan
POST /v1/installer/jobs/{id}/executions    → begin
POST /v1/installer/jobs/{id}/executions/{exec_id}/result → results/outputs
PATCH /v1/installer/jobs/{id}              → status
```

Execution reuses the runner's existing job machinery where possible rather than
reimplementing terraform/helm/manifest execution. The runner's job handlers live
under `bins/runner/internal/jobs/` and `pkg/runner/jobs/`; the air-gap work proved
these can run against a substituted client — it implemented the whole
`nuonrunner.Client` interface against a local store and drove the real job loop.
**Decide early**: link the runner's job packages, or vendor a reduced executor. That
choice drives most of this milestone's size.

### Sources come from the bundle only

All OCI content resolves from the bundle, never from a remote registry. The air-gap
implementation's approach is the right one and worth copying:

- pre-flight the whole workflow and **abort before any infrastructure work** if any
  referenced member is missing from the bundle, with a message naming it;
- fail loudly rather than falling back to a registry.

Related: terraform must **fail closed** when the vendored provider mirror or binary
is missing. The air-gap runner silently fell back to `releases.hashicorp.com`, which
quietly broke the isolation guarantee.

## 3. Approvals

Plan-then-approve is the core customer-control loop and the reason the installer
exists.

1. Execute the plan job, push contents via C2.
2. The control plane creates a `WorkflowStepApproval` and the step awaits.
3. `GET /v1/installer/approvals` → pending approvals; fetch contents.
4. Render in the web UI: the plan diff, affected resources, and — for
   customer-managed components — the underlying template.
5. `POST /v1/installer/approvals/{id}/response` with approve/deny + reason.
6. The conductor resumes or stops.

Reuse the dashboard's plan-diff rendering concepts
(`client/components/approvals/plan-diffs/`) rather than inventing a second diff
format.

**Approval is a mutation**: subject to the CSRF, body-cap, and auth requirements in
04's A1–A6. On a non-loopback bind without authentication, anyone who can reach the
port could approve a production change — which is exactly why A1 refuses to start.

## 4. Status and health push

- Step progress and completion: `PATCH .../steps/{id}/status` and
  `POST .../steps/{id}/complete`, idempotent on `(step_id, attempt)`.
- Job results: the existing `CreateRunnerJobExecutionResult` shape (C2).
- Component health: `POST /v1/installer/component-health`. The runner already has a
  component-health engine (`bins/runner/internal/pkg/componenthealth/`) with
  cluster/terraform/manifest-kinds providers; point its sink at C2.
- Heartbeats: `POST /v1/installer/heartbeats` updates `InstallerAgent.LastSeenAt`.
  The air-gap work deferred heartbeats and then had no liveness signal at all —
  include them.

**Push only what the RFC allows**: workflow status, component health, bundle status.
Not logs by default, not plan contents beyond what an approval needs, and never
secrets.

## 5. Update flow

`installer update` is `install` against a newer bundle:

1. List bundles, identify newer ones (a bundle is 1:1 with an app-branch-run, so
   ordering is by run).
2. Download + verify the new bundle.
3. Fetch the workflow the control plane generated for the update.
4. Diff: which components change, which are added/removed. Show it before anything
   executes.
5. Approve → execute → push.

Two things the air-gap work flagged and did not solve:

- **Bundle-digest handoff**: anything scheduled or in-flight against the old bundle
  must be resolved before switching. Reject work whose bundle digest doesn't match
  the active one rather than executing it.
- **Failed-step retry on resume**: the air-gap runner auto-retried *failed* steps on
  restart with no opt-in, which can re-run a destructive apply. Make retry explicit
  (`--retry-failed`).

## 6. Resumability

The installer will be killed mid-flow. Requirements:

- All C2 writes idempotent (job claim, step completion, results).
- Local progress recorded in the store so a restart resumes rather than restarts.
- If a plan was saved and its apply not yet run, the apply must still be valid on
  resume. The air-gap work hit this concretely: terraform bakes the state-backend
  URL *including the port* into saved plans, so a resumed process had to rebind the
  exact same port — hence `tfbackend-port` persistence. Whatever backend this uses,
  test kill-and-resume between plan and apply.
- **Single writer.** Either enforce mutual exclusion (conditional-write lease on the
  bucket) or document and enforce a single-installer assumption. The air-gap work
  built a careful S3 lease for the multi-writer case and still had no guard against
  two invocations on one *local* state dir.

## Customer-managed resources (RFC, optional for v1)

The RFC proposes `customer-managed: true|false` on components and the sandbox: the
installer shows the underlying template and the customer creates the resource
themselves, pushing outputs back. Three classes — stack, sandbox, component.

This is a substantial feature that touches app config (02), the workflow model
(03), and this flow. **Recommend deferring past v1** and stating so explicitly,
rather than half-implementing it. If it is in v1, it needs its own spec.

## Milestones

| # | Deliverable |
| --- | --- |
| A | Bundle acquisition + fail-closed verification + `InstallBundle` status push |
| B | Job claim/execute/result loop for one component type (terraform), decision made on reusing runner job packages |
| C | Approval loop end to end, rendered in the web UI |
| D | Remaining component types (helm, kubernetes_manifest, external_image) + sandbox |
| E | Component health + heartbeats |
| F | `installer update` incl. bundle-digest handoff and explicit retry |
| G | Resumability: kill/resume between plan and apply, single-writer enforcement |

## Tests

- Verification: good bundle installs; tampered blob, tampered manifest, wrong
  signature, and mismatched transport checksum each abort **before** any execution
  and push `failed`.
- Missing bundle member → pre-flight abort naming the member, with no infrastructure
  touched.
- Approval: deny stops the workflow; approve resumes it; a second response to the
  same approval is rejected.
- Idempotency: replay every C2 write; assert exactly-once effect.
- Kill/resume: SIGKILL between plan and apply, restart, apply completes against the
  saved plan.
- Update: newer bundle produces a correct component diff; a job carrying a stale
  bundle digest is rejected, not executed.
- Terraform fails closed when the vendored mirror/binary is absent.

## End-to-end verification

The air-gap M1 procedure is a good template. Against a sandbox account:

1. Publish a bundle from an app branch run with `bundle_enabled` (02).
2. Create a bundle-mode install (03).
3. `installer connect` + `setup` (04).
4. `installer verify` — confirm all three checks report as *performed*.
5. `installer install` — approve in the web UI, watch it apply.
6. Confirm in the dashboard: workflow steps `success`, component health reported,
   `InstallBundle = applied`.
7. Publish a second bundle, `installer update`, confirm the diff and the re-apply.
8. Tamper with a downloaded bundle and confirm `verify` and `install` both refuse.

## Risks

- **Reusing the runner's job packages vs. a reduced executor** is the biggest
  unknown and drives the size of milestone B. Decide it first; the air-gap work
  showed reuse is feasible via client substitution.
- **The approval surface is a production-mutation surface.** It must inherit 04's
  auth posture; an unauthenticated approval endpoint is worse than no installer.
- **Resumability is where correctness bugs hide** — the air-gap prototype found four
  distinct resume bugs (lost create-plan results, stuck interrupted steps, baked-in
  backend ports, `show -json` overwriting state). Budget for it and test it
  explicitly.
- **Scope**: customer-managed resources and the OAuth server mode are both large.
  Defer both, explicitly.

# 03 — Bundle installs (data model, external execution, UI)

> Owns: bundle-install data model, `external` workflow execution, the installer-facing
> API, and the bundle-install UI. Depends on: nothing. Blocks: 04, 05.
> Contracts: owns **C0**, owns **C2**, owns **C4**.

## Goal

An install can be **customer-managed** (`management_mode = bundle`). Its workflows
are marked `external`: the control plane generates them but does not execute them.
An installer executes them and pushes back a limited set of status — workflow/step
state, job results, component health, bundle state — which the dashboard renders.

## Scope

**In**: the models, the `external` execution axis, the blocking/completion
mechanics, the C2 API surface, the C4 handshake, and the dashboard views.

**Out**: the installer itself (04/05), the bundle format (01), and publishing
(02).

## The three execution axes — read C0 first

The single biggest risk in this spec is overloading an existing enum. Two axes
already exist:

| Axis | Field | Change |
| --- | --- | --- |
| Who **drives the workflow** | `Workflow.ExecutionType` — **new** | `control-plane` (default) \| `external` |
| Who **claims the runner job** | `RunnerJob.Executor` (`internal/app/runner_job.go:233-238`) | add `external` to `org-runner`, `control-plane` |
| What **kind of step** it is | `WorkflowStep.ExecutionType` (`internal/app/workflow_step.go:19-27`) | **unchanged** |

A bundle install's workflow is `external` and *still contains* `approval` steps
(customer approves a plan) and `system` steps (control-plane bookkeeping). One enum
cannot express that. Keep them separate.

`RunnerJob.Executor` already has a non-runner value: `control-plane`, used so orgs
without a runner can build, with `GET /v1/runner-jobs?executor=control-plane`
(`internal/app/runners/service/list_runner_jobs_ctl_api.go:36-64`). The installer's
claim is that shape with a different value — extend it, don't invent a parallel
mechanism.

## Data model

### `Install.ManagementMode`

`managed | bundle`, default `managed`, backfill existing rows to `managed`.
`services/ctl-api/internal/app/install.go:30-161`.

Today "managed" is implicit in three independent things: a `RunnerGroup` (:83), a
sandbox, and a cloud account/connection. The clearest existing predicate is
`shouldCreateManagedAWSCloudFormationStack`
(`installs/signals/awaitinstallstackversionrun/signal.go:74-82`). `ManagementMode`
makes the intent explicit rather than derived — audit every site that infers
managed-ness from `RunnerGroup` presence and decide whether it should now check the
mode.

### `Workflow.ExecutionType`

`services/ctl-api/internal/app/workflow.go:226-295` (table
`install_workflows`). `control-plane | external`, default `control-plane`.

Set at generation time from `Install.ManagementMode`. Step generation lives in
`internal/app/installs/workflows/v2/*.go`; the signal-type → step-metadata mapping
is `getSignalStepMetadata` (`v2/shared.go:105-159`).

### `RunnerJob.Executor` += `external`

`internal/app/runner_job.go:233-238`, defaulted in the hook at :372-373.

### `InstallerAgent` — ID prefix e.g. `iag`

Connected installer CLIs. `InstallID`, `OrgID`, `Status`
(`created | connected | disconnected | revoked`), `BucketConfig` (jsonb),
`Version`, `Fingerprint`, `LastSeenAt`, `ConnectedAt`.

**Named `InstallerAgent`, not `Installer`** — `app.Installer` already exists
(`internal/app/installer.go`): a dead self-hosted-installer model with only a
migration, an org relation (`org.go:160`), and a hard-delete entry
(`orgs/helpers/hard_delete.go:75`). No routes, no services. Dropping or reclaiming
it is separable cleanup.

### `InstallBundle` — per-install bundle state

`InstallID`, `AppConfigBundleID`, `Status`
(`available | downloaded | verified | applying | applied | failed`), `VerifiedAt`,
`AppliedAt`, `StatusDescription`. This is the state the installer pushes.

### `InstallSigningKey` — one active keypair per install

`InstallID`, `PublicKey` (PEM, exposed via API), `PrivateKeyRef` (jsonb, secrets
manager reference, **never returned** — `json:"-"`), `Algorithm` (`ed25519`),
`CreatedAt`, `RotatedAt`, `RevokedAt`.

**Open decision — key custody.** The RFC says the install defines the private key.
Two options with materially different trust stories:

- **Vendor-held** (control plane signs at publish, 02 milestone 5): simple, works
  today, but Nuon holds the signing key.
- **Customer-generated** (customer creates the pair, uploads only the public key):
  stronger, but the vendor cannot sign — the customer would verify against a key
  they control, which changes what the signature *means*.

Resolve before implementing; it changes C1 usage and 02's `sign_bundle`.

## External step mechanics

**Model on `awaitinstallstackversionrun`** — the existing "customer does this out
of band" primitive
(`internal/app/installs/signals/awaitinstallstackversionrun/signal.go`):

1. set the step target,
2. create a `callback.Ref` and persist it,
3. block on `callback.AwaitWithTimeout` (180 days there),
4. completion arrives via an authenticated POST that fires the callback
   (`installs/service/post_install_phone_home.go` → `stackrun` signal →
   `callback.Send` at `stackrun/signal.go:237-274`),
5. on timeout, set a terminal status with a useful message.

It is also the only signal mapped to `WorkflowStepExecutionTypeUser`
(`v2/shared.go:139-141`), and `InstallStackVersionRun.RunType` already has an
`out-of-band-update` value (`install_stack_version_run.go:26-29`) — the notion of
"this state change came from outside the workflow" exists.

### Non-negotiables

- **Status writes go through the existing activity**,
  `statusactivities.AwaitPkgStatusUpdateFlowStepStatus`
  (`internal/pkg/workflows/status/activities/update.go:211`). Do **not** add a
  direct-write path from an HTTP handler — the group reads `ResultDirective` back
  out of the DB to decide what happens next, and bypassing the activity breaks
  that.
- **Metadata merges via `generics.MergeJSONBMetadata`**, never a full status
  rewrite (ctl-api/AGENTS.md). Concurrent installer pushes will otherwise clobber
  each other.
- **Never put `step.Idx` in a user-facing string** — use `step.Name`
  (ctl-api/AGENTS.md).
- **Idempotent completion** on `(step_id, attempt)`. The installer can be killed
  and restarted at any point.

## API (C2)

New service package
`services/ctl-api/internal/app/installer/service/`, registered with
`fx.Provide(api.AsService(installerservice.New))` in
`internal/fxmodules/services.go`. Per ctl-api/AGENTS.md, routes in an unregistered
package silently don't exist — this is the most common way this goes wrong.

Route set is in `contracts.md` C2. Two things to get right:

1. **Token↔path install must match.** Reject any request whose scoped token's
   install differs from the path install. This is stricter than today's runner
   routes, which take `:runner_id` from the path and never cross-check the token
   (`runners/service/get_runner.go:40-52`) — do not copy that.
2. **Reuse existing result shapes.** `CreateRunnerJobExecutionResult`
   (`runners/service/create_runner_job_execution_result.go:83`) already handles
   `Contents`, `ContentsDisplay`, the gzip variants, `ErrorMetadata`, `ErrorCode`.
   Don't invent a second one.

### C4 handshake

`POST /v1/installer-auth/connect` with `{install_id, install_token,
agent_fingerprint, version}` → short-lived scoped token + install metadata + bucket
config.

Build on `POST /v1/installs/{install_id}/runner-bootstrap-token`
(`installs/service/create_runner_bootstrap_token.go:34-68`), which already mints an
install-scoped credential for a customer-side process via
`runners/helpers/token.go` (2h bootstrap / 90d standard, against service account
`<id>@serviceaccount.nuon.co`).

Install tokens must be **revocable**, modelled on
`InstallStackVersion.PhoneHomeTokenRevokedAt` (`install_stack_version.go:45-55`)
and its authorization path `authorizePhoneHome`
(`installs/service/post_install_phone_home_auth.go:73-160`), which enumerates
rejection reasons explicitly — a good pattern to copy.

The control plane **never stores customer cloud credentials**. `bucket_config`
describes *where* state lives; the installer accesses it with its own credentials.

## UI

Three existing dispatchers are the plug-in points — no new rendering framework
needed.

### 1. Step detail panel

`client/components/workflows/step-details/StepDetailPanel/StepDetailPanelContainer.tsx`
— `getStepPanelDetails(step)` (:37-63) switches on `step.step_target_type`. Add an
external-step panel. Also:

- `getStepPanelSize(step)` (:24-35)
- the auto-open heuristic (:150-190), where `isPendingAwaitStack` (:161-165) sits
  beside `isPendingApproval` — **an external step blocked on the customer belongs
  in that pair**
- it polls `getWorkflowStep` every 1.5s until `step_target_type` lands, then 10s
  (:97-105)

### 2. Step action bar

`client/components/workflows/step-details/StepButtons.tsx:13-46` — today renders
skip/retry and approve/deny. This is where any external-step affordance goes.

Note: for a bundle install, "mark this step complete" from the *dashboard* is
questionable — completion should normally arrive from the installer, which has the
actual results. Prefer showing state and instructions (like `AwaitStackDetails`
does) over a vendor-side "mark done" button. If a manual override is wanted, make
it explicitly an override with a reason, not the primary path.

### 3. The await-stack precedent to imitate

`client/components/workflows/step-details/stack-details/` is a working
"customer must do something" UI: `StackStepDetailsContainer` (fetches by
`step.owner_id`, polls every 3s), `AwaitStackDetails` (status card + outputs +
per-cloud panel), `AwaitAWSDetails` (launch link, template, click-to-copy, file
download). An external-step panel should look like this.

### Plus

- A bundle-install install view: workflow statuses, component health, bundle state,
  connected agents.
- `client/lib/ctl-api/` functions following the thin-wrapper pattern in
  `lib/ctl-api/workflows/`.
- Regenerate `client/types/nuon-oapi-v3.d.ts` after the API lands.

## Milestones

1. **Data model**: `ManagementMode`, `Workflow.ExecutionType`,
   `RunnerJob.Executor += external`, `InstallerAgent`, `InstallBundle`,
   `InstallSigningKey` — with migrations and hard-delete ordering.
2. **Generation**: workflows for bundle installs get `ExecutionType = external`
   and their jobs get `Executor = external`. Audit managed-ness inference sites.
3. **External step mechanics**: the blocking signal + callback + idempotent
   completion, modelled on `awaitinstallstackversionrun`.
4. **C4 handshake** + install-token issuance/revocation.
5. **C2 API surface** + SDK generation.
6. **UI**: external-step panel, bundle-install view, agent list.

Milestones 4 and 5 unblock 04, so land them before the installer needs them.

## Tests

- Integration tests per `api-integration-test-builder` conventions for every C2
  route: happy path, wrong-install token → 403, revoked install token → 401,
  cross-org → 404.
- Idempotency: the same step completion posted twice advances the workflow once;
  the same job claimed twice by one agent succeeds, by a second agent fails.
- Temporal test for the external step signal: completion fires the callback and the
  step reaches `success`; timeout produces a terminal status with a message
  containing `step.Name` (not `Idx`).
- Concurrent metadata pushes don't clobber (exercises
  `MergeJSONBMetadata`).
- `ManagementMode = bundle` produces `external` workflows and jobs; `managed`
  produces `control-plane`.

## Risks

- **Enum overloading** (C0). If someone adds `external` to
  `WorkflowStep.ExecutionType` instead of `Workflow`, approvals inside external
  workflows stop working. Call it out in review.
- **Bypassing the status activity.** An HTTP handler writing
  `WorkflowStep.Status` directly will appear to work and will break
  `ResultDirective`-driven group progression. Route everything through the
  activity.
- **Managed-ness is currently inferred in several places.** Adding an explicit mode
  without auditing those sites gives two sources of truth.
- **Key custody is unresolved** and blocks 02's signing milestone.
- **The dead `Installer` model** will cause confusion in code search until it's
  dropped.

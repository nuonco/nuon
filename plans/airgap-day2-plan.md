# Airgap day-2 ops plan (M4–M5)

Goal: demo cron actions, healthchecks, runbooks, and drift checks in an
air-gapped install, with a customer-side portal to observe and trigger
them. Constraints: runner keeps zero ingress and no public egress; the
portal talks only to customer AWS (S3, CloudWatch, ECR); the portal never
generates plans — everything executable is pinned in the bundle.

Oracle-reviewed 2026-08-10. Verdict: feasible. Two reframings and one
constraint correction below are load-bearing.

## Decisions (Harsh, 2026-08-10)

1. Drift-detection network risk accepted: runner sits in the VPC ideally;
   dev allows outside-VPC. The portal runs on the operator machine and
   talks AWS APIs directly.
2. Drift = fresh plan from pinned template (never replayed tfplan).
3. Terraform state goes to S3 via the native S3 backend, replacing the
   loopback HTTP backend + periodic disk→S3 mirror.
4. Runner becomes resident after bootstrap.
5. At-least-once dispatch (requests/claims/receipts) accepted.
6. Self-contained action templates accepted. Additionally: define where
   install state lives offline and add an app-config property that fails
   publish when the config relies on online-only surfaces (see below).
7. Cron semantics (UTC, forbid overlap, skip missed ticks) accepted.
8. Health = component health checks, not just runner self-check. The
   runner componenthealth engine already exists and its providers are
   wired in airgap mode; observations must sink to the state store
   instead of the stubbed API client.

## Reframings

1. **Immutable execution templates, not saved plans.** The bundle pins
   templates (module source, provider mirror, tf binary, inputs, backend
   identity, action step config). Every run instantiates fresh: drift is a
   fresh plan-only run against current state, never a replayed tfplan.
   Drift result means "drifted from the bundle's frozen desired config".
2. **At-least-once dispatch with global serialization.** S3 is a mailbox,
   not a queue. Immutable request objects + conditional claim objects with
   leases + terminal receipts. Duplicate dispatch resolves to the same
   receipt. Actions must be idempotent or key off dispatch_id.

## Constraint correction

"Only S3/ECR/CloudWatch" cannot support real drift: terraform refresh
calls EC2/IAM/EKS/STS APIs. Actual contract: **no public internet;
private/VPC-endpoint access to the AWS APIs the pinned plans use.**
Publish-time output: a network dependency manifest; deploy fails if
required endpoints are missing. Terraform must fail closed when the
vendored mirror/binary is missing (today it falls back to
releases.hashicorp.com — see bins/runner/internal/jobs/deploy/terraform/workspace.go).

## Architecture

```diagram
┌──────────────────────────────┐        ┌─────────────────────────────────┐
│ OPERATOR LAPTOP              │        │ CUSTOMER ACCOUNT                │
│ nuon-bundle portal           │        │                                 │
│  (localhost UI, AWS creds    │        │  S3 state bucket                │
│   from active context)       │        │  ├ dispatch/requests/<ulid>     │
│                              │ writes │  ├ dispatch/claims/<ulid>       │
│  trigger runbook/action ─────┼───────▶│  ├ runs/<run-id>/status,steps   │
│  read runs, logs, results ◀──┼────────┤  ├ runtime/lease.json           │
│                              │        │  ├ schedules/<id>/cursor.json   │
└──────────────────────────────┘        │  └ terraform/<workspace>/...    │
                                        │        ▲          │ poll        │
                                        │        │ write    ▼             │
                                        │  ┌─────┴──────────────────┐     │
                                        │  │ RUNNER VM (resident)   │     │
                                        │  │ local cron scheduler   │     │
                                        │  │ dispatch poller+claims │     │
                                        │  │ jobloop (serialized)   │     │
                                        │  │ templates from bundle  │     │
                                        │  └────────────────────────┘     │
                                        └─────────────────────────────────┘
```

## Envelope v1 (day-2 section)

- Stable IDs + display names; bundle digest + deployment binding.
- **OfflineActionTemplate**: self-contained (resolved step config inline).
  Today's action job fetches run/config from ctl-api and git-clones at
  exec time (bins/runner/internal/jobs/actions/workflow/fetch.go,
  exec_env.go) — v1 accepts only inline commands, no git sources, no
  secret params. Cron schedule string included where configured.
- **Drift template** per terraform component: plan-only op derived from
  the component's composite plan with ApplyPlanContents/PlanJSON cleared;
  runtime op is create-apply-plan. Classify with the existing no-op
  classifier (pkg/plans/types/approval_plan/terraform.go) — resource
  changes, output changes, resource_drift.
- **Component health**: the existing componenthealth engine
  (bins/runner/internal/pkg/componenthealth/, cluster + terraform +
  manifest-kinds providers, already provided in bins/runner/cmd/airgap.go)
  runs resident and reports every 60s. Today the airgap client stubs
  CreateComponentHealth to a silent no-op; replace the stub with a state
  store sink (health/<component>/latest.json + transitions). Portal shows
  per-component health from there. A runner self-check (scheduler
  heartbeat, S3 r/w, job-loop status) is a separate, clearly-labeled row.
- **OfflineRunbookTemplate**: ordered refs to the above, stop on first
  failure. No branching, approvals, retries, event waits, or output
  interpolation in v1. Publish rejects runbooks using unsupported steps.

## Runner: one-shot → resident

Today the airgap run finishes the install envelope, closes status, and
exits (systemd Restart=on-failure). Changes:

- After bootstrap, enter a durable day-2 loop; supervisor stays a process
  manager (ECR login, systemd) only.
- Deployment-wide S3 lease before executing anything (split-brain guard).
- One active run globally (safest for tf state + arbitrary actions).
- Runtime job IDs = run-id + template-step-id; never reuse frozen IDs.
- Per-run layout runs/<run-id>/...; shared terraform state lives under
  terraform/<workspace>/, never under runs/.
- Cron: UTC, five-field parser (same as ctl-api), forbid overlap, skip
  missed ticks (no backfill after outages), record skipped occurrences.
  occurrence_id = hash(deployment_id, bundle_digest, schedule_id,
  scheduled_at_utc), claimed via the same S3 conditional path as manual
  dispatch.
- Terraform state: native S3 backend configured per workspace
  (terraform/<workspace>/), replacing the loopback HTTP backend +
  periodic mirror in pkg/runner/airgap/tfbackend.go and the upload loop
  in bins/runner/cmd/airgap_s3.go. S3 is the single authority; native
  S3 backend lockfile locking guards concurrent runs. Plan chaining
  (create-apply-plan → apply) must keep working against the S3 backend.

## Install state offline + the footgun property

Where install state lives in an air-gapped deployment:

| Online home (ctl-api Postgres/S3)      | Offline home                       |
| -------------------------------------- | ---------------------------------- |
| installs, deploys, workflow steps      | S3 state store runs/ + status.json |
| runner jobs, logs                      | runs/<run-id>/steps/, CloudWatch   |
| component health transitions           | health/<component>/                |
| action runs + step updates             | runs/<run-id>/ (local synthesis)   |
| terraform state (Nuon-managed S3)      | customer S3, native backend        |
| install inputs (non-secret)            | inputs.yaml → S3, late-bound       |
| secrets                                | NOT AVAILABLE (deferred)           |
| approvals, webhooks, Slack, event bus  | NOT AVAILABLE                      |

State under a deployment prefix has explicit ownership:

```text
state/
├── runner/   # status, runs, logs, plans, health, tfstate, receipts
└── control/  # dispatch requests, install controls, candidates, approvals
```

The runner keeps only runner-owned durable files in its local state directory
and mirrors that entire directory to `runner/`. There is no artifact allowlist:
a new runner artifact is durable without a second sync configuration change.
The portal writes only through the `control/` store and reads an overlay of
`runner/`, `control/`, and the legacy flat layout. A successor runner restores
legacy runner-owned files once, then lets namespaced files win. `LEASE` and
`DONE` remain transport-level coordination keys at the deployment state root
for stack-bootstrap compatibility; they are never part of bulk state sync.

Footgun today: the airgap API client silently no-ops many write surfaces,
so vendor configs can appear to work while state goes nowhere.
(DONE 2026-08-10 for component health: CreateComponentHealth persists
snapshots to health/latest.json and appends health/transitions.json;
health context survives restarts via health/context.json;
GetRunnerInstallComponents serves component metadata frozen into the
envelope at export.)

PUNTED (2026-08-10): the app-config `airgap` compatibility declaration
(publish-time blocking qualification, fail-closed client stubs, capability
matrix surface). Revisit after M4.

## Install inputs offline (non-secret)

The envelope already exports the input schema (InputSpec: name, type,
description, secret, required — pkg/runner/airgap/envelope.go) and
`inspect` shows it. Missing: the values path. Plans are rendered with the
reference install's input values baked in, so customer values must
replace them, like stack outputs.

Design (chosen: publish-time placeholder rewrite):

1. Export: plan_rewrite.go rewrites baked input values into unique tokens
   (`__NUON_INPUT_<name>__`). Collision/ambiguity surfaces at publish in
   the qualification report, not at deploy. Export input defaults too
   (InputSpec.Default currently missing).
2. Customer: `inputs:` in the init file or inputs.yaml; validated against
   specs (required/type) at `stack prepare`; uploaded to S3 next to stack
   outputs.
3. Runner: latebind substitutes placeholder → customer value; any
   leftover placeholder fails loudly as a missing required input.
4. Sensitive inputs excluded: publish fails/warns if a sensitive value is
   baked into plans; no offline secret path yet.

IMPLEMENTED (2026-08-10), on top of PR 2097:

- export: plan_inputs.go rewrites reference values → `__NUON_INPUT_<name>__`
  (substring for values ≥6 chars, exact leaf/comma-segment for 3–5,
  never for shorter or true/false); InputSpec gains Default + Bindable;
  secret values baked into plans fail publish; duplicate reference
  values fail publish.
- runner: `--install-inputs` (path or s3://), validated then substituted
  at render; leftover placeholders fail the step with the input names.
- CLI: `stack prepare --inputs inputs.yaml` validates, uploads
  config/inputs.json, wires AIRGAP_INSTALL_INPUTS_URI through init
  script → airgapmng → runner service; `inspect` INPUTS table shows
  DEFAULT + OFFLINE (editable / secret / fixed at publish) and decodes
  hex override input names to `override:<kind>:<component>` (accepted
  back as aliases in the inputs file).

Known limits: inputs whose reference value is short/generic (bools,
small numbers) are not late-bindable; ambiguous shared values block
publish; no offline secrets.

## Dispatch protocol

- Portal writes dispatch/requests/<dispatch-id>.json with If-None-Match:*
  {schema_version, deployment_id, bundle_digest, ref_id, dispatch_id,
  requested_by (display only, not audit truth), created_at}.
- Runner validates ref exists in the bundle envelope; conditionally
  creates dispatch/claims/<dispatch-id>.json (runner generation, lease
  expiry, attempt, run ID); terminal receipt written only after run state
  is durable. Stale claims retryable after lease timeout.
- IAM: portal principals write only requests/ and read state; runner
  reads requests/ and writes claims/runs/state. CloudTrail S3 data events
  for authoritative caller identity.
- **No params in v1.** M5: vendor-declared typed non-secret params that
  fill dedicated value slots only (never command text, sources, job
  type/op, backend, auth, deps). No secrets in dispatch docs ever.

## Portal (`nuon-bundle portal`)

Localhost only: bind 127.0.0.1 random port, creds stay in the Go backend,
no CORS, strict Host/Origin (DNS rebinding), CSRF token on mutations,
same-origin assets. Read-only inspect of refs/schedules; trigger pinned
refs; run history + live status + S3-backed logs/results. No editing, no
CloudWatch tail, no cancellation, no RBAC UI in v1. In-cluster portal
rejected as first cut (needs healthy cluster + ingress/TLS/auth).

## M4 (~3–6 days) — demo slice

1. Envelope v1 + publish-time rejection of unsupported templates.
   (DONE 2026-08-10: envelope carries optional Actions/Drift/Runbooks
   with validation + lookup helpers; ctl-api export dedupes action
   templates, rejects Git-backed rendered actions, derives drift
   templates from raw plan JSON, rewrites late-bound input placeholders
   across bootstrap/actions/drift; cron Git actions block
   qualification, non-cron Git actions warn + are excluded; bundle
   creation accepts runbooks through the publish signal.)
2. Resident runtime, run model, lease, serialized execution.
   (DONE 2026-08-10: run model + serialized day-2 execution landed —
   `day2` package defines the S3 contract (dispatch requests/claims/
   receipts, runs, schedules, catalog); `day2run` has the serialized
   executor (actions, drift with fresh-plan classification, runbooks
   with stop-on-failure + health gates), S3 mailbox with conditional
   claims/takeover, 15s poller with digest/ref validation + rejected
   receipts, UTC cron scheduler with deterministic occurrence IDs and
   busy-skip cursors, catalog publication; wired into the resident
   block of `runner airgap` when --state-s3 is set; run artifacts flush
   to S3 before the receipt is written.)
   (Earlier note 2026-08-10: `runner airgap` always stays resident after
   a successful bootstrap with the componenthealth engine running (no
   flag; the systemd unit runs the same command); the S3 `LEASE` object with
   conditional writes (If-None-Match/If-Match, 90s TTL, 30s renewal)
   guards the deployment; `nuon-bundle health [--follow]` reads
   health/latest.json + transitions. Run model + serialized day-2
   execution still open.
   Lease hardening after Oracle review: per-write 15s timeouts;
   freshness anchored to write attempt start; loss declared 45s after
   the last confirmed write; takeover of an expired-looking lease
   requires a 165s (ttl + fail-stop grace + one renewal) ETag-stability
   watch — clock-independent proof the holder is dead or has
   force-exited; a loser's final state upload/DONE/release are gated on
   an ownership context canceled at loss, and the process force-exits
   after 60s if shutdown hangs; a resident runner retries DONE
   publication; local LEASE/DONE files are never bulk-uploaded.
   Residual risk, accepted for M4: cancellation is cooperative — an
   in-flight terraform apply/helm child process can outlive the
   fail-stop exit. True fencing (process-group kill of children,
   fencing tokens on downstream writes) is M5 work.)
3. Dispatch + cron per above.
   (DONE 2026-08-10: see item 2; `nuon-bundle refs/run/runs` drive the
   mailbox with If-None-Match dispatch creation and claim/run/receipt
   polling.)
4. Minimal portal.
   (DONE 2026-08-10: `nuon-bundle portal` — localhost-only embedded UI,
   Host-header + CSRF checks, dependency-free page refreshing every 5s;
   lists refs/runs, triggers dispatches, shows drift verdicts. Never
   constructs a plan; only writes dispatch requests.)
5. Demo: cron inline action every few minutes; component health (kill a
   deployed workload pod → unhealthy → recovers); runbook = health gate →
   action; drift on one terraform component (out-of-band change →
   drifted; revert → no-drift). All in the portal.

Acceptance: install finishes and runner stays resident; reboot loses no
accepted dispatch and replays no completed one; concurrent triggers
serialize; duplicate dispatch maps to same receipt; drift runs produce
fresh plan JSON; no public route and terraform/provider/git fallbacks
fail closed; portal never constructs a plan.

## M5 — hardening

- S3-authoritative tf state (conditional writes) replacing disk→S3
  mirror; distributed locks; ASG replacement + stale-lease recovery.
- Schedule catch-up/coalescing policy; retention/compaction.
- Typed non-secret params; vendored git action sources; declared
  tool/image deps by digest.
- Per-ref IAM authz; CloudTrail-backed audit identity.
- CloudWatch live tail; cancellation.
- Bundle upgrade: schedule handoff between bundle digests.
- Drift classifiers for Helm/manifest/Pulumi; application-health
  templates distinct from runner self-check.

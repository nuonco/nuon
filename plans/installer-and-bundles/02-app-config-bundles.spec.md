# 02 — App-config bundles (publish side)

> Owns: `AppConfigBundle`, app config `bundle_enabled`, the branch-run publish
> step, the Bundles UI. Depends on: **C1** (01). Blocks: nothing.
> Contracts: consumes C1, fills **C3**.

## Goal

An app opts into bundles in its config. Every app-branch-run then publishes one
`AppConfigBundle` — a signed, checksummed artifact containing every component,
the sandbox, and an encoded set of workflows. The dashboard lists them.

## Scope

**In**: the model, config plumbing, the branch-run step and publish signal, the
workflows manifest (C3), signing per bundle-mode install, upload + verify, and the
app-level Bundles page.

**Out**: install-side state (`InstallBundle`, `InstallSigningKey` → 03), the
installer (04/05), and the bundle format itself (01).

## Data model

New models in `services/ctl-api/internal/app/app_config_bundle.go`. Register ID
domains in `pkg/shortid/domains/domains.go`, migrations in
`services/ctl-api/internal/pkg/db/psql/models.go`, and hard-delete ordering in
`services/ctl-api/internal/app/orgs/helpers/hard_delete.go` (children before
parent).

### `AppConfigBundle` — ID prefix `acb`

| Field | Notes |
| --- | --- |
| `ID`, `CreatedByID`, `CreatedAt` | standard; `BeforeCreate` fills from context |
| `OrgID`, `AppID`, `AppConfigID` | scoping |
| `AppBranchRunID` | **unique index** — the RFC's 1:1 with an app-branch-run |
| `SandboxBuildID`, `ComponentBuildIDs map[string]string` | jsonb, keyed by component config connection ID — the pinned builds |
| `TargetPlatform`, `SchemaVersion` | |
| `ManifestDigest`, `TransportChecksum`, `Size` | content identity |
| `WorkflowsManifest` | jsonb, the C3 workflows document |
| `Status` | `pending \| publishing \| active \| error` |
| `StatusDescription` | |
| `Artifacts []AppConfigBundleArtifact` | `OnDelete:RESTRICT` |
| `Replicas []AppConfigBundleReplica` | `OnDelete:RESTRICT` |

### `AppConfigBundleArtifact` — ID prefix `aba`

One row per bundle member, mirroring the OCI manifest into SQL so the API can list
contents without opening the archive. Unique on `(bundle_id, kind, logical_name)`.
Fields: `Kind`, `LogicalName`, `ComponentConfigConnectionID`, `ConfigDigest`,
`SourceType`, `SourceIdentity` (jsonb), `Repository`, `Digest`, `MediaType`,
`Size`, `PlatformOS`, `PlatformArchitecture`.

### `AppConfigBundleReplica` — ID prefix `abr`

A stored copy. `Provider`, `Region`, `StorageRef` (`json:"-"` — the storage key
never leaves the server), `StorageVersion`, `TransportChecksum`, `Size`,
`VerifiedAt *time.Time`.

**A bundle is downloadable only when `Status == active` AND it has a replica with
`VerifiedAt != nil`.** Keep "we published" and "we read the bytes back and they
matched" as separate facts.

### Uniqueness caveat

`AppBranchRunID` unique gives real protection against double-publish, which the
air-gap prototype lacked (it keyed on `manifest_digest`, only known *after*
publishing, so concurrent creates raced). Add the unique index, and still make the
publish activity idempotent.

## Config plumbing — `bundle_enabled`

Per the RFC this is an **app** field.

1. **Config schema** — add to `config.AppConfig` (`pkg/config/config.go:10`, the
   root struct) plus a `JSONSchemaExtend` entry (~:79). Parsing is generic
   mapstructure decode (`pkg/config/parse/`), so no parser work.
2. **Model** — `App.BundleEnabled bool` (`services/ctl-api/internal/app/app.go:27`).
   Note `app.App` has **no boolean fields at all** today; this is the first. Put it
   with the config-derived scalars at :35-38 or beside
   `ConfigRepo`/`ConfigDirectory` at :81-82. Needs a migration.
3. **Sync** — written in
   `services/ctl-api/internal/pkg/config/syncer/app_metadata.go`. Add to **both**
   the `updates` literal and the `Select(...)` column list (~:19-28); omitting the
   `Select` entry silently drops the write.

### Deviation to resolve

`syncApp` writes `App` columns directly — there is no `internal/pkg/config/build/app.go`.
AGENTS.md says config must become models through `internal/pkg/config/build`. Either
add `build/app.go` and route through it (correct, slightly more work) or note the
existing deviation explicitly in the PR. Do not silently follow the bad pattern.

Cover the field in `internal/pkg/config/build/build_test.go` — per AGENTS.md a
dropped config field is invisible without a test.

## Branch-run publish step

`AppBranchRun` (`services/ctl-api/internal/app/apps/workflows/app_branch_run.go:26`)
builds sequential step groups:

| Order | Step | Line |
| --- | --- | --- |
| 1 | `setup preview` (git_preview only) | :52 |
| 2 | `fetch commit` / skipped | :64 / :85 |
| 3 | `fetch app config` / skipped | :74 / :92 |
| 4 | `building components and sandbox` / skipped | :109 / :102 |
| — | **preview runs return here** | :121 |
| 5 | per-install-group `plan` / `deploy` / post-deploy runbooks | :137+ |

**Insert `publish bundle` between step 4 and the preview early-return (~:117-119).**
Bundles need the pinned builds from step 4 and previews should not publish.

Follow the append pattern at :108-117 exactly:

```go
sg.nextGroup()
step, err := sg.appBranchSignalStep(ctx, appBranchID, "publish bundle", pgtype.Hstore{}, &publishbundle.Signal{
    AppBranchID: appBranchID,
    RunID:       runID,
}, WithSkippable(false))
if err != nil { return nil, errors.Wrap(err, "unable to create publish bundle step") }
steps = append(steps, step)
```

Helpers: `appBranchSignalStep` (`apps/workflows/shared.go:104`), `nextGroup` (:63),
options `WithSkippable` (:16) / `WithExecutionType` (:43).

Gate on `app.BundleEnabled`. When false, either skip the group entirely or emit a
skipped variant matching the existing `fetch commit (skipped)` convention at :85.

## Publish signal

New package `services/ctl-api/internal/app/apps/signals/branches/publishbundle/`,
mirroring the sibling `builds/`. Register in the app-queue signal catalog.

Activities, in order:

1. **`get_bundle_inputs`** — resolve the run's pinned `ComponentBuild`s (by
   `AppBranchRunID`) and the `AppSandboxBuild`. Require `active` status on each;
   fail with a precondition error naming what's missing.
2. **`create_bundle_record`** — insert `AppConfigBundle` (`status = publishing`).
   Idempotent on `AppBranchRunID`.
3. **`encode_workflows`** — snapshot workflows, runbooks, and actions into the C3
   workflows manifest. **This is the highest-risk activity in this spec** — see
   below.
4. **`publish_bundle`** — assemble members via C1, build the archive, upload via
   `Transport`, verify read-back.
5. **`sign_bundle`** — for each bundle-mode install of this app, sign the manifest
   digest with that install's key (03 owns `InstallSigningKey`; this activity
   consumes it). Skip cleanly when there are no bundle-mode installs.
6. **`update_bundle_status`** — `active`, or `error` with the message.

Retry policy: bounded (3 attempts). Start-to-close generous — the air-gap
equivalent needed 180m for a ~374 MB bundle.

### Idempotency

The air-gap prototype's publish was **not** idempotent: re-publishing hit the
artifact unique index and exhausted every Temporal retry. Fix by construction —
delete the bundle's artifact rows and re-insert **in the same transaction** as the
status update, and short-circuit when digests are already set and a replica is
verified.

### Encoding workflows — the hard part

`encode_workflows` must turn live config into a frozen, self-contained execution
template. Constraints learned from the air-gap work:

- **Reject, don't degrade.** If a runbook or action uses something that cannot be
  frozen (git-sourced action steps, event waits, output interpolation, secret
  params), **fail the publish with a clear message** naming the offending
  workflow. Silently excluding it produces a bundle that looks complete and isn't.
- **No secrets, ever.** If a secret input's value appears in any encoded plan, fail
  publish. This must be a hard error, not a warning.
- **Per C3, `pkg/bundle` must not parse this.** The manifest carries a
  `runbook:<name>` *member*; the semantics live in the workflows document.

Produce a qualification-style report of violations and warnings, return it in the
API error body on failure, and store it on the bundle for the UI.

## API

Public routes, org+app scoped, `@Security APIKey` + `@Security OrgID`. New service
package `services/ctl-api/internal/app/app-config-bundles/service/`, registered via
`fx.Provide(api.AsService(...))` in `internal/fxmodules/services.go` — per
ctl-api/AGENTS.md, routes in a package that isn't registered there simply don't
exist.

```
GET  /v1/apps/{app_id}/bundles                        list, paginated, created_at DESC
GET  /v1/apps/{app_id}/bundles/{bundle_id}            detail + artifacts
POST /v1/apps/{app_id}/bundles/{bundle_id}/download-grants   presigned download
```

Responses omit `OrgID`, `StorageRef`, and build IDs. The download grant returns
`{url, expires_at, filename, size, transport_checksum, manifest_digest,
supports_range}` — the client needs the digests to verify, and per C1/S1 the CLI
must actually compare them.

Publishing is triggered by the branch run, not by an endpoint. A manual
`POST /v1/apps/{app_id}/bundles` re-publish endpoint is optional; if added, it must
reuse the same signal and idempotency.

## UI

- **Page**: `services/dashboard-ui/client/views/app/Bundles.tsx`, route in
  `views/app/routes.tsx` (children of `<AppLayout />`, path shape
  `':orgId/apps/:appId/bundles'`).
- **Nav**: `views/app/AppLayout.tsx` — add to `navLinks` (~:79-108) as
  `{ path: '/bundles', iconVariant: '...' as const, text: 'Bundles' }`, gated
  `app?.bundle_enabled && {...}` with the existing `.filter(Boolean) as TNavItem[]`
  at :108. Follow `hasAppBranchesUI` (:81).
  **Caveat**: `AppTemplate` early-returns for `hasNewAppIA` (:68-75) and renders no
  SubNav — the new-IA path needs its own entry or the tab won't appear there.
- **Components**: `client/components/apps/bundles/{BundlesTable,BundleDetail}/`
  following the existing `Component`/`Container`/`stories` triplet.
- **API client**: `client/lib/ctl-api/apps/bundles/` — thin `api({method, orgId,
  path})` wrappers matching `lib/ctl-api/workflows/`.
- Surface `status`, `manifest_digest` (shortened, click-to-copy), `size`,
  `created_at`, the branch run link, artifact count, and **signature state**.
  Per C1/S5, do not label anything "signed" in the UI until signing is wired.
- The download action must surface the digest. The air-gap dashboard did
  `window.location.assign(grant.url)` and never showed one, making the browser path
  strictly weaker than the CLI's.

## Milestones

1. Model + migrations + ID domains + hard-delete ordering.
2. `bundle_enabled` through config → sync → model, with a `build_test.go` case.
3. Publish signal + activities against a **stubbed** C1 builder (unblocks before 01
   lands): record, status transitions, idempotency.
4. `encode_workflows` + the qualification report.
5. Real C1 integration: pack, upload, verify, sign.
6. API + UI.

## Tests

- Integration tests for the three endpoints per the
  `api-integration-test-builder` conventions: pagination, org scoping (a bundle
  from another org must 404), grant refused when not `active` or no verified
  replica.
- Publish signal: happy path status transitions; **re-publish is idempotent**;
  missing component build → precondition error; git-sourced action step → publish
  fails with the workflow named; secret value in an encoded plan → publish fails.
- `syncer` test asserting `bundle_enabled` round-trips (config → DB), plus the
  `build_test.go` case.
- Branch-run step generation: `bundle_enabled` true inserts the group in the right
  position; false does not; a preview run never publishes.

## Risks

- **`Select(...)` omission** in `app_metadata.go` silently drops the write. Cover
  with a test, not a code read.
- **`encode_workflows` scope creep.** It is tempting to support every workflow
  feature. Ship a narrow, explicitly-rejecting v1 — the air-gap work proved that
  the failure mode of silent degradation is much worse than an upfront error.
- **Signing depends on 03** for `InstallSigningKey`. Sequence milestone 5 after
  that model lands, or stub the key lookup behind an interface.
- **Bundle size.** The air-gap equivalent was 374 MB, dominated by the terraform
  provider mirror. Confirm the storage bucket, lifecycle policy, and per-org growth
  before enabling this broadly.

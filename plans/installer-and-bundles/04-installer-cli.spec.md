# 04 — `installer-cli` (scaffolding, setup, installer-api)

> Owns: `bins/installer`. Depends on: **C1** (01), **C2**/**C4** (03).
> Blocks: 05. Same binary and probably the same owner as 05.

## Goal

`bins/installer` — a customer-run CLI that connects to an install, provisions its
own state bucket in the customer's cloud, and runs an **installer-api** with an
embedded web UI. This spec covers everything up to "the installer is running and
can read/write install state." The actual install/update flow is 05.

## Scope

**In**: cobra scaffolding, `connect`/`setup`, bucket provisioning, the local
state store, the installer-api (local + remote modes), the embedded web UI, and the
customization surface.

**Out**: executing workflows and the approval flow (05), the bundle format (01),
control-plane models (03).

## Naming

Binary `installer`, invoked as `nuon-installer` per the RFC. Package
`bins/installer`. The control-plane row representing a connected instance is
`InstallerAgent` (03).

## Milestone A — scaffolding

Follow `bins/cli`, not `bins/runner`. `bins/runner` uses fx dependency injection,
which is right for a long-lived service and wrong for a customer-facing CLI.

Structure:

```
bins/installer/
  main.go
  cmd/
    root.go          cobra root, persistent flags
    cli.go           the cli struct + persistentPreRunE
    connect.go       C4 handshake
    setup.go         bucket provisioning
    run.go           installer-api + web UI
    status.go verify.go bundles.go
    annotations.go   output-mode declaration
  internal/
    config/          context file handling
    api/             the installer-api server
    store/           blob-backed state store
    cloud/            per-cloud bucket provisioning
    ui/              output helpers
  web/               embedded SPA
```

Reuse from `bins/cli`:

- `--output table|json|agent` machinery: `cmd/annotations.go` (`OutputTable`/
  `OutputJSON`/`OutputAgent`, `outputsAnnotation`), `internal/ui/agent.go` (the
  `{ok, data, error}` envelope, `emitAgentSuccess`/`emitAgentError`/`classifyError`),
  `internal/agentmode/`. Resolution precedence is in `cmd/cli.go:70`
  (`resolveOutput`).
- Config pattern from `internal/config/config.go` (YAML file, env overlay,
  `BindCobraFlags`).
- `internal/ui/print.go` for `CLIUserError` and friends.

Context/config file: `~/.config/nuon-installer/contexts/<name>` with a `current`
pointer, written `0600`. Multiple contexts so one operator can manage several
installs; `installer ctx` to switch. Never store the install token in plaintext if
an OS keychain is available; if it is not, `0600` and say so in the docs.

### Build and release — verify before starting

**There is no Makefile and no goreleaser config anywhere in the repo.** Confirm how
a new binary is built, versioned, and released in CI (look at `actions/` and
`scripts/`) *before* assuming a new `bins/` entry ships. This CLI is also
customer-distributed, so it needs published checksums at minimum, and per the RFC it
should be OSS-buildable by the vendor.

## Milestone B — `connect` and `setup`

### `installer connect`

```
installer connect --install-id <id> --install-token <token>
```

C4: `POST /v1/installer-auth/connect` → scoped token + org/app/install metadata +
`management_mode` + bucket config. Persist to the context; refresh the scoped token
by re-presenting the install token. Handle revocation with a clear message telling
the operator to get a new token from the vendor.

### `installer setup` — bucket provisioning

**This is the largest genuinely-new piece of work in the whole plan.**

Nothing in the repo creates a bucket in any cloud. An exhaustive grep for
`CreateBucket`, `AWS::S3::Bucket`, `aws_s3_bucket`, `google_storage_bucket`,
`azurerm_storage_account`, and `Microsoft.Storage/storageAccounts` finds only
comments and test fixtures. Every bucket in use today is pre-existing and
config-supplied.

What *does* exist and is worth extending: the three-cloud template-rendering
machinery under `services/ctl-api/internal/pkg/stacks/`:

| Dir | Cloud | Entry point |
| --- | --- | --- |
| `cloudformation/` | AWS | `cloudformation.go:9` `Templates`, `:19` `NewTemplates` |
| `arm/` | Azure | `arm.go:15` `Templates`, `:38` `Template(inp)` |
| `aws/` | AWS (terraform) | `render.go:71` `Render(...)` |
| `gcp/` | GCP (terraform) | `render.go:90` `Render(...)` |

These are mature and well-tested for roles, secrets, VPCs, and phone-home. Adding a
bucket resource is a natural extension. Note the lifecycle difference: those render
an *install stack* the customer runs to provision a runner; `setup` provisions a
bucket before anything else exists.

Also note two gaps: there is **no unified multi-cloud credential interface**
(`pkg/{aws,gcp,azure}/credentials` share a naming convention but no interface), and
**no Azure blob dependency** (`azblob` is absent from `go.mod`;
`BlobStorageProvider` validation is `oneof=s3 gcs` at
`services/ctl-api/internal/config.go:555`).

**Scope the first cut to AWS.** Define the `cloud.Provisioner` interface for all
three and stub GCP/Azure with clear "not implemented" errors.

```go
type Provisioner interface {
    Provision(ctx context.Context, spec BucketSpec) (bundle.ObjectRef, error)
    Validate(ctx context.Context, ref bundle.ObjectRef) error
}
```

`setup` must, non-negotiably:

1. **Enable versioning** — the C1 transport pins object versions and the air-gap
   publish path required it.
2. **Enable default encryption (SSE)** — terraform state and plan contents land
   here. The air-gap prototype set SSE on *nothing*.
3. **Enable public-access-block** and fail loudly if it can't.
4. **Scope IAM narrowly** — prefix-condition `ListBucket`, no account-wide grants.
   The air-gap runner policy collapsed to `<bucket>/*` when the prefix was empty and
   granted account-wide ECR writes.
5. **`Validate` is idempotent and re-runnable** — verify an existing bucket meets
   all of the above rather than only creating a new one.

## Milestone C — installer-api + web UI

A local HTTP server with two modes (RFC):

- **local** — all state read from and written to the customer bucket. No control
  plane.
- **remote** — proxies the control plane with the C4 scoped token.

It's a backend-for-frontend over a pluggable store. Define one interface and two
backends:

```go
type Store interface {
    Get(ctx context.Context, key string, out any) error
    Put(ctx context.Context, key string, v any) error
    List(ctx context.Context, prefix string) ([]string, error)
}
```

The reference `pkg/runner/airgap/statestore` on `982ff3ced` is *shaped* right
(object-key-per-concern) but has only a disk backend, and its `LockTF`/`UnlockTF`
are filesystem operations. **Object stores need conditional-write/precondition
semantics for locking, and nothing in the repo does that today.** If the installer
needs mutual exclusion, design it on conditional writes (`If-None-Match` /
`If-Match`) explicitly.

### Patterns to follow

- `services/bundle-portal/server.go` on `982ff3ced` — the closest existing thing:
  embedded React via `go:embed`, Go 1.22 method-pattern routes, host allowlist +
  CSRF, `writeJSON`/`writeAPIError` helpers, a committed `dist/` so no JS toolchain
  is needed at build time. Also `main.go`'s `allowedHosts` / `requestHost`
  (DNS-rebinding protection).
- `pkg/runner/airgap/tfbackend.go` on `982ff3ced` — loopback HTTP over a `Store`
  interface, including **port persistence** across restarts.
- `services/dashboard-ui/server/internal/spa/serve.go` — the production
  SPA-with-history-fallback pattern.

### Security — designed in, not discovered

The air-gap portal shipped each of these wrong. They are acceptance criteria.

| # | Requirement | What went wrong |
| --- | --- | --- |
| A1 | **Bind `127.0.0.1` on a random port by default.** Any non-loopback bind requires real authentication — bearer token or mTLS — and refuses to start without it | `services/bundle-portal` supported a non-loopback bind with **no authentication at all** |
| A2 | **The CSRF token is not a credential.** Do not inject it into the served HTML and treat it as auth | It was injected into `index.html`, so anyone who could `GET /` could read it and dispatch |
| A3 | Host allowlist + `Origin` check (DNS rebinding), `nosniff`, `X-Frame-Options: DENY`, tight CSP with `frame-ancestors 'none'` | the standalone service had these; the CLI-embedded variant did not |
| A4 | Body size cap + `DisallowUnknownFields` on every mutation; validate all path params | only the standalone service did this |
| A5 | Cloud credentials stay in the Go process, never reach the browser | (the prototype got this right — keep it) |
| A6 | No unvalidated passthrough of stored JSON into responses | `p.health` string-concatenated raw stored JSON |

If the RFC's "run it on a server with an OAUTH HTTP API" is in scope, it is a
separate, explicitly-authenticated mode — not a flag on the loopback server.

### `installer verify`

Per C1/S1, verification **fails closed**. `verify` takes the expected digest (from
the C2 download grant or `--manifest-digest`/`--checksum`), calls
`bundle.VerifyBundle`, and reports exactly which checks ran. It must never print
`ok` for a check it didn't perform — the air-gap `verify` printed
`transport checksum: ok` unconditionally with nothing to compare against, while its
README claimed the bundle was "signed" when nothing was.

## Customization (RFC)

The installer is meant to be vendor-brandable and OSS-buildable:

- A config file (YAML or JSON) shipped alongside the binary: product name, logo,
  support links, which views are enabled, theme tokens.
- A stylesheet override for the embedded UI.
- Design the config surface in this milestone even if the theming implementation
  lands later, so the file format doesn't have to change once vendors depend on it.

## Milestones

| # | Deliverable |
| --- | --- |
| A | Cobra scaffolding, context/config handling, output modes, CI build path confirmed |
| B1 | `connect` against C4; token refresh + revocation handling |
| B2 | `setup` for AWS with versioning/SSE/public-access-block/scoped IAM, idempotent `Validate` |
| B3 | `Provisioner` interface + GCP/Azure stubs |
| C1 | `Store` interface + bucket backend + control-plane backend |
| C2 | installer-api routes with the A1–A6 security posture |
| C3 | Embedded web UI: read-only views (bundles, workflows, health) |
| C4 | `verify` with fail-closed semantics |

## Tests

- Unit: context file round-trip and `0600` permissions; output-mode resolution
  precedence; `allowedHosts` including the wildcard-bind refusal.
- `setup` against a sandbox AWS account: creates a compliant bucket; re-running
  validates and does not error; a bucket missing versioning or SSE fails
  `Validate` with a specific message.
- installer-api: non-loopback bind without auth **refuses to start**; CSRF absent →
  rejected; bad `Origin` → rejected; oversized body → rejected; unknown JSON field
  → rejected.
- `verify`: good bundle passes; tampered blob fails; tampered manifest fails; wrong
  signing key fails; a missing `--manifest-digest` reports the check as *skipped*
  rather than passed.
- Local vs remote store parity: the same view renders from either backend.

## Risks

- **Bucket provisioning is net-new** across three clouds, with no repo precedent
  and no unified credential abstraction. It is the schedule risk in this plan.
  AWS-only first.
- **Azure needs a new module** (`azblob`).
- **Object-store locking** has no precedent here. If concurrent installers are
  possible, design conditional-write leases deliberately; if not, state the
  single-writer assumption and enforce it.
- **The non-loopback mode is the highest-severity security surface** in the whole
  plan. Defaulting to loopback and refusing unauthenticated wildcard binds is the
  single most valuable guardrail.
- **Two CLIs drifting**: the air-gap work had `bins/nuon-bundle/cmd/portal.go` as a
  near-duplicate of `services/bundle-portal/server.go`, and the CLI copy was missing
  the security headers, body cap, and param validation. Share one implementation.

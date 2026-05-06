# Composite Errors

Status: draft
Owner: TBD

A platform-wide abstraction for capturing, classifying, persisting, and rendering
errors that arise across the Nuon platform — workflow steps, runner jobs, component
builds, install deploys, and the components that drive them.

Today, errors flow up to the user as a single `status_description` line on a step
or as raw bytes inside a runner job execution result. There is no way to:

- attach rich, type-safe context to an error (e.g. the missing IAM action and the
  resource ARN behind a Terraform apply failure);
- drive workflow control-flow off an error (fail fast vs. retry vs. downgrade to
  warning);
- render a useful error view on a deployment object that pulls together logs,
  parsed output, and remediation hints;
- collect statistics about what kinds of errors orgs are hitting.

Composite Errors are the abstraction that fixes that. They are first-class,
catalogable, persisted entities. They mirror the existing **signal catalog** pattern
(`pkg/queue/catalog`) — types are registered in-process via `init()`, hydrated from
JSONB at read time, and behave like typed Go values everywhere they are consumed.

## Goals

- A typed, catalogable error abstraction that can be **persisted, attached to many
  owner types, rendered in the dashboard, and queried in admin tooling**.
- A **parser pipeline** that classifies raw error material (Temporal errors, runner
  stderr, terraform output, helm output, …) into typed errors.
- The ability for an error to **override the directive** the conductor would
  otherwise apply (fail-fast, force-retry, skip-group, downgrade-to-warning).
- A **tree of causes** so an `install_deploy_failed` can point at a
  `terraform_apply_failed` whose primary cause is `aws_missing_iam_permission`.
- A path to **multi-error per owner** so a single step can carry several related
  errors when relevant.

## Non-goals (v1)

- API error envelopes — `stderr.Err` already does that job.
- Cross-owner deduplication (e.g. one `aws_missing_iam_permission` row attached to
  12 steps). Out of scope; we will revisit if grouping at read time isn't enough.
- Org-level overrides (admins customizing severity / behavior per error type).
- Runner-side classification. v1 keeps the runner dumb; ctl-api parses everything.
- Localization.

## Design overview

```diagram
╭──────────────────────────────╮  init() registers   ╭──────────────────────────────╮
│ Catalog (in-memory)          │◀────────────────────│ Per-package error type       │
│  type → factory()            │                     │ + parser(s) + Render()       │
╰──────────────┬───────────────╯                     ╰──────────────────────────────╯
               │
               ▼
╭──────────────────────────────╮  polymorphic owner   ╭──────────────────────────────╮
│ composite_errors (DB row)    │◀────────────────────│ workflow_step                │
│  - type, domain, severity    │                      │ runner_job_execution_result  │
│  - data jsonb (typed)        │                      │ component_build              │
│  - source jsonb (snippet)    │                      │ install_deploy / install     │
│  - references jsonb          │                      │ ...                          │
│  - title/summary cached      │                      ╰──────────────────────────────╯
│  - resolution fields         │
╰──────────────┬───────────────╯
               │
               ▼
╭──────────────────────────────╮      ╭──────────────────────────────╮
│ composite_error_causes       │      │ Conductor / step finalizer   │
│  parent_id, child_id,        │      │  hydrates typed instance,    │
│  idx, is_primary             │      │  honors OverrideDirective()  │
╰──────────────────────────────╯      ╰──────────────────────────────╯
```

The model has two distinct layers:

1. **Type layer (in-process):** Go interface + capability interfaces. Each error
   type lives in its own package under `internal/app/composite_errors/types/<name>/`
   with `error.go`, `parser.go`, and `init.go` calling `catalog.Register(...)`.
   This mirrors `internal/app/installs/signals/v2/<name>/`.
2. **Instance layer (DB):** one `composite_errors` row per recorded incident, with
   a JSONB `data` column that round-trips through the catalog factory back into
   the typed Go value.

## The two axes: parse context vs. error domain

It is important to distinguish *where the error material came from* from *what kind
of error it is*. They use different shapes and they cross-cut.

| Axis | Lives on | Hierarchical | Examples |
|---|---|---|---|
| **`ParseContext`** — *where the bytes came from* | parser registration + dispatch input | yes (path-like) | `terraform`, `terraform/plan`, `terraform/apply`, `helm/install`, `helm/install/rbac`, `kubernetes/rollout`, `runner/job`, `build/helm` |
| **`Domain`** — *what kind of error this is* | the error row + type | no (flat enum) | `terraform`, `helm`, `kubernetes`, `aws`, `gcp`, `azure`, `nuon`, `runner` |

`Domain` lives on the row for filtering / metrics / UI grouping. `ParseContext` is
purely a dispatch concept — it never gets stored.

The split is what lets `aws_missing_iam_permission` (a single parser) run on
TF-apply, Helm-install, and runner-job output without being re-registered into a TF
hierarchy.

## Type interface

```go
// pkg/composite_error/error.go
package composite_error

type Type     string
type Domain   string
type Severity string

const (
    SeverityFatal   Severity = "fatal"
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityInfo    Severity = "info"
)

// CompositeError is the in-memory typed view of a stored composite error.
// Implementations live in internal/app/composite_errors/types/<name>/error.go
// and are registered into the catalog via init().
type CompositeError interface {
    Type() Type
    Domain() Domain
    Severity() Severity

    // Render produces the user-facing view from the typed Data on the
    // implementing struct. Always called at read time — never cached
    // anywhere except the denormalized title/summary on the row, which
    // exist purely for admin search/listing.
    Render(ctx context.Context) Render
}

type Render struct {
    Title       string           // one line, ≤120 chars
    Summary     string           // one paragraph, user-facing
    Sections    []RenderSection  // ordered: what / why / how to fix
    UserActions []UserAction     // CTAs (copy IAM JSON, run command, retry, link)
}

type RenderSection struct {
    Heading string
    Body    string  // markdown
}

type UserAction struct {
    Kind  UserActionKind  // link | copy | command | retry
    Label string
    Value string          // url, snippet, command
}
```

### Optional capability interfaces

Following the `signal/interfaces.go` pattern: each capability is a tiny opt-in
interface that the conductor / helpers check at runtime via type assertion.

```go
// Single override knob — replaces a swarm of per-capability interfaces.
// Returning the zero value means "no opinion, defer to signal defaults".
type ErrorWithDirective interface {
    OverrideDirective() Directive
}

type Directive struct {
    Kind DirectiveKind  // continue | stop | retry | retry-group | skip-group | await-approval
    // Future-proofed: leave room without breaking the interface
    // MaxRetries *int
    // Backoff    *time.Duration
}

// Optional rendering / linkage helpers
type ErrorWithDocsLink   interface{ DocsURL() string }
type ErrorWithReferences interface{ References() []Reference }
type ErrorWithJSONSchema interface{ JSONSchema() []byte } // validate Data on Record
```

The `Directive.Kind` vocabulary aligns with `WorkflowStep.ResultDirective` so the
conductor can apply it without translation.

The conductor merges directives in this order:

```diagram
╭──────────────────────╮     ╭──────────────────────────╮
│ signal defaults      │ →  │ error catalog override    │
│ (AutoRetry,          │     │ (OverrideDirective() if   │
│  MaxRetry, etc.)     │     │  the type implements it)  │
╰──────────────────────╯     ╰──────────────────────────╯
```

Org-level overrides may be added in a future iteration without changing the
interface — they slot in after catalog overrides.

## Catalog (in-memory, mirrors signal catalog)

```go
// pkg/composite_error/catalog/catalog.go
package catalog

var typeRegistry = map[composite_error.Type]func() composite_error.CompositeError{}

func Register(typ composite_error.Type, fn func() composite_error.CompositeError) {
    if _, exists := typeRegistry[typ]; exists {
        panic(fmt.Sprintf("duplicate composite error type registered: %q", typ))
    }
    typeRegistry[typ] = fn
}

// Hydrate looks up the factory for `typ`, instantiates an empty value,
// and unmarshals `data` (the JSONB blob) into it.
func Hydrate(typ composite_error.Type, data []byte) (composite_error.CompositeError, error) { ... }
```

Each error type's package looks exactly like a signal package today:

```
internal/app/composite_errors/types/aws_missing_iam_permission/
├── error.go      // typed struct + Type()/Domain()/Severity()/Render()/OverrideDirective()
├── parser.go     // implements composite_error.Parser
├── init.go       // catalog.Register(Type, factory) + parser_catalog.Register(parser)
└── parser_test.go
```

## Parsers

```go
// pkg/composite_error/parser.go
package composite_error

type ParseContext string  // path-like, "/" separator, e.g. "terraform/plan"

type Parser interface {
    // Contexts this parser opts into. Each entry matches itself + descendants.
    //   {"terraform"}        → terraform, terraform/plan, terraform/apply
    //   {"terraform/plan"}   → only terraform/plan and below
    //   {"terraform","helm"} → cross-cutting (e.g. AWS perm parser)
    Contexts() []ParseContext

    Parse(ctx context.Context, in ParseInput) ParseResult
}

type ParseInput struct {
    Raw         []byte                       // logs / stderr / stdout / JSON output
    ExitCode    int
    GoErr       error                        // raw Go/Temporal error (use HumanError to clean)
    Ctx         ParseInvocationContext       // owner type/id, component type, install id, etc.
}

type ParseResult struct {
    Matched  bool
    Error    CompositeError                  // typed instance, hydratable
    Causes   []ParseResult                   // optional cause sub-results
    Source   Source                          // small input snippet to persist
    Refs     []Reference                     // dynamic refs (log_stream, plan, etc.)
}
```

Empty `Contexts()` is a registration error — every parser must opt in to at least
one subtree.

### Dispatch rule

The pipeline runs in the conductor / step finalizer. Given a `ParseContext` like
`terraform/plan`:

1. Walk ancestors most-specific → least-specific:
   `terraform/plan` → `terraform` → `""` (root).
2. At each level, run every registered parser in registration order.
3. **First non-empty match at the most-specific level becomes the primary error.**
4. Other matches at any level are kept and attached as **secondary errors**
   (separate rows, polymorphically owned by the same owner).
5. If nothing matches, emit a single `unknown_error` primary.

Example: a Terraform plan failing on missing IAM permission produces:

```
primary:   terraform_plan_failed     (matched by terraform/plan parser)
   └─ cause: aws_missing_iam_permission   (matched by AWS parser at "terraform")
```

Parser registration order is deterministic by Go's `init()` ordering, but tests
should not depend on it across packages. For parsers at the same context level, we
sort by `(depth desc, registration order)`.

### Parsing guarantees

- Parsers are **best-effort** — a panicking or erroring parser must never break the
  workflow. The pipeline recovers panics and logs.
- Parsers must be **fast** (microsecond-millisecond). They run on the failure path
  of every step.
- `unknown_error` is always the safety net.

## Storage

### `composite_errors`

```go
type CompositeError struct {
    ID            string                     `gorm:"primaryKey;check:id_checker,char_length(id)=26"`
    CreatedByID   string
    CreatedBy     Account                    `json:"-"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     soft_delete.DeletedAt

    OrgID         string                     `gorm:"index;not null"`     // RLS
    Org           Org                        `json:"-"`

    // polymorphic owner — same pattern as QueueSignal
    OwnerID       string                     `gorm:"index:idx_ce_owner;not null"`
    OwnerType     string                     `gorm:"index:idx_ce_owner;not null"`

    // catalog classification
    Type          composite_error.Type       `gorm:"index;not null"`
    Domain        composite_error.Domain     `gorm:"index;not null"`
    Severity      composite_error.Severity   `gorm:"index;not null"`

    // typed payload — round-trips through Catalog.Hydrate(Type, Data)
    SchemaVersion int                        `gorm:"not null;default:1"`
    Data          datatypes.JSON             `gorm:"type:jsonb;not null"`

    // small parser-input snippet (capped) so we can debug parser decisions later
    Source        Source                     `gorm:"type:jsonb"`

    // references resolved at read time (log streams, plan results, …)
    References    References                 `gorm:"type:jsonb"`

    // denormalized for admin search / list — populated at write time from Render()
    TitleCached   string                     `gorm:"type:text"`
    SummaryCached string                     `gorm:"type:text"`

    // resolution lifecycle (separate from row existence)
    ResolvedAt    *time.Time
    ResolvedByID  *string
    ResolvedNote  string                     `gorm:"type:text"`
}

type Source struct {
    ParserName    string `json:"parser_name,omitempty"`
    ParserVersion string `json:"parser_version,omitempty"`
    Snippet       string `json:"snippet,omitempty"`        // ≤ 8 KB cap
    ExitCode      *int   `json:"exit_code,omitempty"`
    GoError       string `json:"go_error,omitempty"`        // HumanError() output
}

type References []Reference
type Reference struct {
    Type  ReferenceType  `json:"type"`            // log_stream | runner_job_execution_result | terraform_plan_result | doc_url | ...
    ID    string         `json:"id,omitempty"`    // entity id, or url for *_url types
    Label string         `json:"label,omitempty"`
    Meta  map[string]any `json:"meta,omitempty"`  // e.g. {"start_line":1421,"end_line":1438}
}
```

Indexes:

- `(org_id)` for RLS scans.
- `(owner_type, owner_id, deleted_at)` for the owner-side preload.
- `(type)`, `(domain)`, `(severity)` for admin filtering / metrics rollups.
- `(created_at desc)` for recent-errors views.

### `composite_error_causes`

A separate edge table keeps the cause graph queryable in both directions and
multi-cause-friendly.

```go
type CompositeErrorCause struct {
    ParentID  string    `gorm:"primaryKey;not null"`  // higher-level error
    ChildID   string    `gorm:"primaryKey;not null"`  // underlying cause
    Idx       int       `gorm:"not null"`             // ordering of children
    IsPrimary bool      `gorm:"not null"`             // the headline cause
    CreatedAt time.Time
}
```

Constraints:

- FKs on both sides into `composite_errors(id)` with `ON DELETE CASCADE`.
- Unique `(parent_id, child_id)`.
- Helper enforces no cycles before insert.
- Exactly one row per parent may have `IsPrimary = true` (enforced in helper +
  partial unique index).

### Owner-side change

Each owner that wants to display attached errors declares a single GORM
polymorphic association — no FK migration on the owner.

```go
// WorkflowStep
CompositeErrors []CompositeError `gorm:"polymorphic:Owner;polymorphicValue:install_workflow_steps" json:"composite_errors,omitempty"`

// RunnerJobExecutionResult
CompositeErrors []CompositeError `gorm:"polymorphic:Owner;polymorphicValue:runner_job_execution_results" json:"composite_errors,omitempty"`

// ComponentBuild, InstallDeploy, InstallWorkflow, Install, Component, ...
```

There is no central registry of allowed owner types — anything can attach errors,
mirroring how anything can own a `QueueSignal`. This keeps the abstraction generic.

### Soft-delete and retention

- `composite_errors` rows soft-delete with their owner. When the owner is hard
  deleted, errors cascade.
- Resolution does **not** delete a row — `ResolvedAt` is set and the row stays
  queryable for analytics ("this org keeps hitting throttling").
- A future retention sweeper hard-deletes resolved errors older than N days.

## Helpers

Following the helpers pattern in `ctl-api/AGENTS.md`, all cross-domain access
goes through `internal/app/composite_errors/helpers/`.

```go
type Helpers struct {
    cfg *internal.Config
    db  *gorm.DB
    l   *zap.Logger
}

// RecordInput is the unified write entrypoint.
type RecordInput struct {
    OwnerID, OwnerType string
    Error              composite_error.CompositeError  // typed instance
    Source             composite_error.Source
    References         []composite_error.Reference
    Causes             []RecordCause                   // child errors to record + link
}

type RecordCause struct {
    Error      composite_error.CompositeError
    Source     composite_error.Source
    References []composite_error.Reference
    IsPrimary  bool
}

// Record persists the error (and its causes recursively) in a single transaction.
// Validates Data against ErrorWithJSONSchema if implemented.
// Renders Title/Summary into the cached columns.
func (h *Helpers) Record(ctx context.Context, in RecordInput) (*app.CompositeError, error)

// RecordFromError is the convenience entry point used by step finalizers.
// It runs the parser pipeline against the raw inputs and records the result.
func (h *Helpers) RecordFromError(
    ctx context.Context,
    owner Ownerable,
    parseCtx composite_error.ParseContext,
    in composite_error.ParseInput,
) (*app.CompositeError, error)

// Hydrate loads the row and returns the typed instance via Catalog.Hydrate.
func (h *Helpers) Hydrate(ctx context.Context, id string) (*app.CompositeError, composite_error.CompositeError, error)

// ListByOwner returns all attached errors for an owner, ordered by
// (severity desc, created_at asc). Used for the dashboard / public API.
func (h *Helpers) ListByOwner(ctx context.Context, ownerType, ownerID string) ([]*app.CompositeError, error)

// Primary returns the single highest-severity, oldest error attached to an
// owner, or nil. Convenience for badge / status surfaces.
func (h *Helpers) Primary(ctx context.Context, ownerType, ownerID string) (*app.CompositeError, error)

// Tree returns the cause graph rooted at id, depth-bounded.
func (h *Helpers) Tree(ctx context.Context, id string, maxDepth int) (*ErrorTree, error)

// Resolve marks a row resolved (does not delete).
func (h *Helpers) Resolve(ctx context.Context, id, byAccountID, note string) error
```

The helpers package is registered via FX in `cmd/cli.go` next to the existing
`accountshelpers.New` etc.

## Lifecycle: write path

```diagram
                                                 ╭─ catalog override ─╮
   step error          parsers (context tree)    │  Directive.Kind     │
        │                       │                ╰──────────┬──────────╯
        ▼                       ▼                           │
╭───────────────╮    ╭──────────────────────╮               │
│ HumanError()  │──▶│ Pipeline.Parse(ctx,   │──▶ typed CE ──┤
│ on Temporal   │    │   ParseInput)         │   + causes    │
│ error chain   │    ╰──────────────────────╯               │
╰───────────────╯                                            ▼
                              ╭──────────────────────────────────────╮
                              │ helpers.Record():                    │
                              │  - validate (JSONSchema)             │
                              │  - render → Title/Summary cached     │
                              │  - persist row + edges (one tx)      │
                              │  - attach polymorphic owner          │
                              ╰──────────────────────────────────────╯
                                                 │
                                                 ▼
                              ╭──────────────────────────────────────╮
                              │ conductor merges                     │
                              │  signal defaults ⊕ catalog override  │
                              │  → step ResultDirective              │
                              ╰──────────────────────────────────────╯
```

Where parsing runs:

| Surface | Runs in |
|---|---|
| Workflow step failures | conductor / step finalizer activity in ctl-api |
| Runner job execution results | runner-result intake handler in ctl-api |
| Component build failures | build completion path in ctl-api |
| Install deploy aggregations | deploy finalizer in ctl-api |

Runner stays "dumb" in v1 — it reports raw `RunnerJobExecutionResult` and ctl-api
parses. We may push parsing into the runner later for faster cancellation; the
abstraction does not preclude this.

## Lifecycle: read path

- **Public API** — owner endpoints (step, build, deploy, …) preload
  `CompositeErrors` via the polymorphic association and include them in the
  response. Frontend renders by `Type` using the typed `Data`.
- **Dashboard UI** — generic `<CompositeError />` component reads `Title` /
  `Summary` / `Sections` / `UserActions` and dereferences `References` lazily
  (e.g. fetches the linked `LogStream`). Type-specific subviews can be registered
  in a small frontend registry; falls back to generic markdown render.
- **Admin Dashboard** — two new pages under `internal/app/admin-dashboard/`:
  - **Catalog browser**: tree of `ParseContext` nodes with parsers and the error
    types they can emit. Mirrors a Signals catalog browser.
  - **Instance search**: filter by org / type / domain / severity / owner /
    resolved.
- **CLI** — surfaces `Title` + first `UserAction` in step output;
  `--verbose` prints the full structured form.

## Severity vs. step status

- Severity is a property of the error, not a new step status.
- A `warning`-severity error attached to a successful step does not change the
  step status. The dashboard renders a badge sourced from
  `count(composite_errors WHERE severity='warning')`.
- A `fatal` or `error` severity composite error on a failed step is the headline
  surface — it replaces the long form `status_description` in the UI. The short
  form `status_description` stays as a one-liner for compact rendering.
- The conductor only acts on `OverrideDirective()`; severity itself does not move
  the workflow.

## Warnings during successful execution

Warnings can be emitted mid-execution against successful owners:

```go
helpers.RecordWarning(ctx, owner, parseCtx, parseInput)
```

This is the same code path as `RecordFromError` but only produces a row when a
matched parser yields severity `warning`. Keeps "rich event" decoupled from
"failure handling".

## Resolution semantics

- **Implicit:** when an owner that has open errors transitions to a terminal
  success state on retry, the helpers mark all open errors on that owner
  `resolved` automatically (`ResolvedNote = "auto-resolved by retry success"`,
  `ResolvedByID = nil`).
- **Explicit:** an admin or user can call the resolve endpoint with a note
  (e.g. "added IAM permission to role X").
- Resolved errors stay queryable forever (until the retention sweeper clears
  them). They power historical analytics and admin dashboards.

## Initial type catalog

Start small; let parsers drive the roadmap.

| Type | Domain | Severity | Notes |
|---|---|---|---|
| `unknown_error` | `nuon` | `error` | Always-last fallback. |
| `nuon_internal_error` | `nuon` | `error` | DB write fails, transient infra issues. Directive: `retry`. |
| `terraform_init_failed` | `terraform` | `error` | |
| `terraform_plan_failed` | `terraform` | `error` | |
| `terraform_apply_failed` | `terraform` | `error` | |
| `terraform_state_locked` | `terraform` | `error` | Directive: `retry` with backoff once that's wired in. |
| `aws_missing_iam_permission` | `aws` | `error` | Directive: `stop`. Parses action + resource ARN out of stderr. |
| `aws_quota_exceeded` | `aws` | `error` | Directive: `stop`. |
| `aws_throttled` | `aws` | `warning`/`error` | Directive: `retry`. |
| `helm_install_failed` | `helm` | `error` | |
| `kubernetes_resource_not_ready` | `kubernetes` | `error` | |
| `oci_artifact_pull_failed` | `nuon` | `error` | Image not present in customer ECR. |
| `runner_unhealthy` | `runner` | `error` | Heartbeat / dispatch failure. |
| `component_no_active_build` | `nuon` | `error` | Sync-time validation. Directive: `stop`. |

Each type lives in its own package under
`services/ctl-api/internal/app/composite_errors/types/<name>/`.

## Phased delivery

### Phase 0 — design lock-in
This document. No code.

### Phase 1 — core abstraction
- `pkg/composite_error/`: interface, capability interfaces, `Type` / `Domain` /
  `Severity` enums, `Render`, `Reference`, `Source`, `Parser`, dispatch
  pipeline, `unknown_error` builtin.
- `pkg/composite_error/catalog/`: in-memory registry (mirrors `pkg/queue/catalog`).
- `app.CompositeError` + `app.CompositeErrorCause` GORM models + migration.
- `composite_errors/helpers/`: `Record`, `RecordFromError`, `Hydrate`,
  `ListByOwner`, `Primary`, `Tree`, `Resolve`.
- Round-trip tests (encode → store → hydrate → render).
- Pipeline tests against synthetic parsers.

### Phase 2 — workflow step integration
- Add `CompositeErrors` polymorphic association to `WorkflowStep`.
- Hook step-completion path: wrap `HumanError()`, run pipeline, persist via helper.
- Apply `OverrideDirective()` in the conductor.
- Preload + return `composite_errors` from public step endpoints.
- Generic dashboard renderer (no per-type UI yet).

### Phase 3 — first parsers (highest user payoff)
- `terraform_init_failed`, `terraform_plan_failed`, `terraform_apply_failed`,
  `terraform_state_locked`.
- `aws_missing_iam_permission`, `aws_quota_exceeded`, `aws_throttled`.
- Parser fixtures in `testdata/` per package, captured from real failures.

### Phase 4 — broader integration
- Add associations + recording to `RunnerJobExecutionResult`, `ComponentBuild`,
  `InstallDeploy`, `Install`.
- Helm + Kubernetes parsers.
- Cause chaining wired from runner-job → step → deploy.

### Phase 5 — UX
- Per-type renderers in dashboard-ui (start with `aws_missing_iam_permission`).
- Admin dashboard catalog browser + instance search.
- CLI surface.

### Phase 6 — operations
- Implicit auto-resolve on retry success.
- Retention sweeper.
- Metrics: `composite_errors_recorded{type,severity,domain}` counter,
  `composite_errors_unknown` ratio (drives parser coverage roadmap).

## Open follow-ups (not blocking v1)

- **Cross-owner attachment / dedup.** One row attached to many owners. Defer; revisit if read-time grouping by `(type, data hash)` isn't enough.
- **Org-level overrides.** Admins customizing severity / behavior per error type.
- **Runner-side classification.** Push parsers into the runner for fast local fail-fast and reduced ctl-api round-trips.
- **i18n.** `i18nKey` on `Render` outputs.
- **Richer `Directive`.** Grow into a struct with `MaxRetries *int` and `Backoff *time.Duration` if/when parser knowledge demonstrably beats signal defaults.

## Related

- [`pkg/queue/signal`](file:///Users/prem/space/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal) — reference pattern for catalog + capability interfaces.
- [`pkg/queue/catalog/catalog.go`](file:///Users/prem/space/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog/catalog.go) — in-memory registry pattern.
- [`internal/app/status.go`](file:///Users/prem/space/nuonco/nuon/services/ctl-api/internal/app/status.go) — `CompositeStatus` shape (JSONB + `Scan`/`Value`/`GormDataType`).
- [`internal/app/queue_signal.go`](file:///Users/prem/space/nuonco/nuon/services/ctl-api/internal/app/queue_signal.go) — polymorphic owner pattern.
- [`pkg/queue/signal/human_error.go`](file:///Users/prem/space/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/human_error.go) — Temporal error unwrapping; consumed by the parser pipeline.

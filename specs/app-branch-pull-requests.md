# App Branch Pull Requests

Status: design. No code in this pass — see `client/components/playground/pull-request/`
for the UI proof-of-concept that goes with it.

## Problem

A pull request that triggers Nuon previews has no object representing it. Preview runs
(`AppBranchRun`) hang off an `AppBranch` and carry `PRNumber` in a column and in
`Metadata`. Three things fall out of that:

1. **Configuration is per-run, not per-PR.** `AppBranchRunPreview.OverridePreviewConfig`
   lives on the run, so choosing an install or a mode applies to exactly one run and is
   lost on the next push.
2. **The sticky PR comment has no home.** `AppBranchRun.GithubCommentID` is per-run, so
   `find_previous_run_comment_id.go` has to walk backwards over runs to find the comment
   to edit.
3. **There is no "all runs on this PR" view.** Listing runs for a PR means filtering a
   branch's runs by `pr_number`. There is nowhere to hold PR title, state, draft status,
   or head/base refs, and nowhere to render the per-commit run history as one thing.

## Goal

- The first comment on a new PR links into the dashboard, where you configure an install
  and a preview mode **for that PR**.
- Every later visit to that page shows the PR's run + commit history across all preview
  runs on the PR.

## Identity

A pull request is scoped to **`(app_branch_id, number)`**, unique.

A PR belongs to the `AppBranch` whose name matches the PR's base ref. This is the same
scope the sticky comment marker already uses
(`<!-- nuon-app-branch-preview:{app_branch_id} -->`, see `pr_comment_marker` in
`create_or_update_pr_comment.go`), and the same scope run ownership already uses.

One GitHub PR targeting a repo watched by two app branches produces two rows. Cross-app
fan-out into a single object is explicitly not modeled — see *Out of scope*.

## Model

New file `services/ctl-api/internal/app/app_branch_pull_request.go`.

```go
type AppBranchPullRequestState string

const (
    AppBranchPullRequestStateOpen   AppBranchPullRequestState = "open"
    AppBranchPullRequestStateClosed AppBranchPullRequestState = "closed"
    AppBranchPullRequestStateMerged AppBranchPullRequestState = "merged"
)

// AppBranchPullRequestConfigStatus is derived, never hand-set. See "Config status".
type AppBranchPullRequestConfigStatus string

const (
    AppBranchPullRequestConfigStatusReadyFromDefaults AppBranchPullRequestConfigStatus = "ready-from-defaults"
    AppBranchPullRequestConfigStatusNeedsInput        AppBranchPullRequestConfigStatus = "needs-input"
    AppBranchPullRequestConfigStatusConfigured        AppBranchPullRequestConfigStatus = "configured"
)

type AppBranchPullRequest struct {
    ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero"`
    CreatedByID string                `gorm:"not null;default:null" json:"created_by_id,omitzero"`
    CreatedBy   Account               `json:"-"`
    CreatedAt   time.Time             `json:"created_at,omitzero"`
    UpdatedAt   time.Time             `json:"updated_at,omitzero"`
    DeletedAt   soft_delete.DeletedAt `json:"-"`

    OrgID string `gorm:"notnull;default null" json:"org_id,omitzero"`
    Org   Org    `faker:"-" json:"-"`

    // uniqueIndex(app_branch_id, number)
    AppBranchID string    `gorm:"not null" json:"app_branch_id,omitzero"`
    AppBranch   AppBranch `faker:"-" json:"app_branch,omitempty"`
    Number      int       `gorm:"not null" json:"number,omitzero"`

    // GitHub provenance, refreshed on every pull_request event.
    Title       string `json:"title,omitempty"`
    AuthorLogin string `json:"author_login,omitempty"`
    HTMLURL     string `json:"html_url,omitempty"`
    RepoOwner   string `json:"repo_owner,omitempty"`
    RepoName    string `json:"repo_name,omitempty"`
    HeadRef     string `json:"head_ref,omitempty"`
    BaseRef     string `json:"base_ref,omitempty"`
    HeadSHA     string `json:"head_sha,omitempty"`

    State   AppBranchPullRequestState `gorm:"not null;default:'open'" json:"state,omitzero"`
    IsDraft bool                      `json:"is_draft,omitempty"`

    VCSConnectionID *string        `json:"vcs_connection_id,omitempty"`
    VCSConnection   *VCSConnection `json:"-"`

    OpenedAt *time.Time `json:"opened_at,omitempty"`
    ClosedAt *time.Time `json:"closed_at,omitempty"`
    MergedAt *time.Time `json:"merged_at,omitempty"`

    // Nuon scope — hoisted off the run.
    ConfigStatus          AppBranchPullRequestConfigStatus `gorm:"not null;default:'ready-from-defaults'" json:"config_status,omitzero"`
    PreviewOverride       *AppBranchPreviewOverride        `gorm:"type:jsonb;serializer:json;default:null" json:"preview_override,omitempty"`
    ResolvedPreviewConfig AppBranchPreviewConfig           `gorm:"type:jsonb;serializer:json;not null;default:'{}'" json:"resolved_preview_config,omitzero"`
    GithubCommentID       *int64                           `json:"github_comment_id,omitempty"`

    Runs      []AppBranchRun `gorm:"foreignKey:AppBranchPullRequestID" json:"runs,omitempty"`
    LatestRun *AppBranchRun  `gorm:"-" json:"latest_run,omitempty"`
    RunCount  int            `gorm:"-" json:"run_count,omitzero"`
}
```

Indexes (via the `Indexes(db *gorm.DB) []migrations.Index` method, matching
`AppBranchRunPreview`): `org_id`, `app_branch_id`, and a unique
`(app_branch_id, number)`. `BeforeCreate` assigns
`domains.NewAppBranchPullRequestID()` and fills `CreatedByID` / `OrgID` from context, as
every other model in `internal/app` does.

### Changes to existing models

**`AppBranchRun`** (`internal/app/app_branch_run.go`) gains:

```go
AppBranchPullRequestID *string               `json:"app_branch_pull_request_id,omitempty"`
AppBranchPullRequest   *AppBranchPullRequest `json:"app_branch_pull_request,omitempty"`
```

`PRNumber` and `Metadata.PRNumber` stay as-is for compatibility; the FK is backfilled
from them. `GithubCommentID` keeps being written per run, but the PR row becomes
authoritative: `find_previous_run_comment_id.go` reads
`AppBranchPullRequest.GithubCommentID` and only falls back to the run walk for rows
predating the backfill.

**`AppBranchPreviewOverride`** (`internal/app/app_branch_preview_config.go`) is today
`{mode, install_id}`. Widen it so a PR-scoped override can express everything a branch
default can:

```go
type AppBranchPreviewOverride struct {
    Mode          *AppBranchRunPreviewMode `json:"mode,omitempty"`
    InstallID     *string                  `json:"install_id,omitempty"`
    InstallName   *string                  `json:"install_name,omitempty"`
    LabelSelector *labels.Selector         `json:"label_selector,omitempty"`

    SetStatuses  *bool `json:"set_statuses,omitempty"`
    Comment      *bool `json:"comment,omitempty"`
    IgnoreDrafts *bool `json:"ignore_drafts,omitempty"`
    React        *bool `json:"react,omitempty"`
}
```

Pointer fields throughout, so "unset" is distinguishable from "set to false" — the same
reason `AppBranchPreviewConfig.UnmarshalJSON` already uses `*bool` on its wire type.

**`AppBranchRunPreview`** is unchanged. It stays the immutable per-run snapshot:
`BranchPreviewConfig` / `OverridePreviewConfig` / `ResolvedPreviewConfig` still record
what *that* run used, which is what makes the history table meaningful when the PR's
config changes mid-stream.

## Config resolution

A run's `AppBranchRunPreview.ResolvedPreviewConfig` is layered:

```
AppBranchConfig.PreviewConfig                   branch default
  ← AppBranchPullRequest.PreviewOverride        PR scope, set from the UI      [NEW]
  ← TriggerAppBranchRunRequest.PreviewRun       per-run, from CLI / manual
```

Only the middle layer is new. `AppBranchPullRequest.ResolvedPreviewConfig` caches the
first two layers merged, so the UI and the comment can show what the *next* run will do
without simulating a trigger. Everything downstream — `setuppreview`,
`planinstallgroup`, `updateinstallgroup` — is untouched.

### Config status

`ConfigStatus` is recomputed whenever the branch config version changes or the PR
override is written. It is a function of the existing
`AppBranchPreviewConfig.Validate()`:

| Value | Condition | Comment CTA |
| --- | --- | --- |
| `configured` | `PreviewOverride != nil` | "Configure this preview →" |
| `ready-from-defaults` | no override, and `AppBranchConfig.PreviewConfig.Validate() == nil` — the defaults name an install / name / label selector, or mode is `build-only` | "Configure this preview →" |
| `needs-input` | no override, and `Validate()` fails — a non-`build-only` mode with no install target | "**Action required** — choose an install →" |

This is the product rule: *if your preview settings already work, they are used; if they
do not, you pick an install.* It reuses `Validate()` rather than restating the condition,
so the two cannot drift.

## Lifecycle

`internal/app/vcs/signals/github_event/execute.go` upserts the `AppBranchPullRequest`
before it decides whether to trigger a run. Today only `opened`, `synchronize`, and
`ready_for_review` are handled; the rest are logged and dropped.

| Action | Effect |
| --- | --- |
| `opened` | create the PR row, compute `ConfigStatus`, post the comment, trigger a run |
| `synchronize` | update `HeadSHA`, trigger a run with the resolved config |
| `ready_for_review` | `IsDraft = false` (already triggers a run) |
| `converted_to_draft` | `IsDraft = true` — **new**, row update only |
| `edited` | refresh `Title` — **new**, row update only |
| `reopened` | `State = open` — **new**, row update only |
| `closed` | `State = merged` if the payload's `merged` is true, else `closed`; stop triggering — **new** |

The four new actions update the PR row and do not trigger runs. Run-trigger gating in
`matchesRunConfig` (`internal/app/vcs/worker/activities/`) is unchanged. No teardown on
`closed`: previews target pre-existing installs.

### First run

The first run always fires using the branch defaults — nothing blocks on someone
visiting the dashboard. When `ConfigStatus == needs-input` there is no install to target,
so the run degrades to `build-only`: the config is validated and components are built,
which requires no install. The comment then leads with the action-required CTA. Saving a
PR override in the UI offers a re-run at the configured mode.

## API

Mounted in `internal/app/apps/service/service.go` beside the existing
`.../branches/:app_branch_id/runs` routes, gated by `OrgFeatureAppBranches`.

```
GET   /v1/apps/:app_id/branches/:app_branch_id/pull-requests
GET   /v1/apps/:app_id/branches/:app_branch_id/pull-requests/:number
PATCH /v1/apps/:app_id/branches/:app_branch_id/pull-requests/:number/preview-config
GET   /v1/apps/:app_id/branches/:app_branch_id/pull-requests/:number/runs
POST  /v1/apps/:app_id/branches/:app_branch_id/pull-requests/:number/runs
GET   /v1/pull-requests
```

- `GET :number` embeds `resolved_preview_config`, `config_status`, `latest_run`, and
  `run_count`.
- `PATCH .../preview-config` takes an `AppBranchPreviewOverride` body, validates the
  merged result with `AppBranchPreviewConfig.Validate()`, writes `PreviewOverride`,
  recomputes `ResolvedPreviewConfig` and `ConfigStatus`, and re-renders the sticky
  comment.
- `POST .../runs` re-runs the PR at its current resolved config — a thin wrapper over
  `TriggerAppBranchRun` with `PreviewRun.Source = pr` and the PR's number / head SHA.
- `GET .../runs` preloads `Preview`, `VCSConnectionCommit`, and `Comparison`. That join
  is exactly what the run history table renders.
- `GET /v1/pull-requests` is org-wide, mirroring the existing `GET /v1/branches`.

Install candidates for the picker come from the existing
`GET .../branches/:app_branch_id/preview-install-candidates`.

## Comment

`pr_comment_body.go` / `create_or_update_pr_comment.go`:

- `PRCommentParams` gains `PRURL` and `ConfigStatus`.
- The primary CTA becomes the PR page
  (`/{orgId}/apps/{appId}/branches/{branchId}/pull-requests/{number}`). The existing
  per-run "View preview run →" link stays below it.
- When `ConfigStatus == needs-input`, a `> [!IMPORTANT]` block with the configure link
  replaces the install-impact section.
- A run-history line, "*N preview runs on this PR*", links to the same page.

The marker stays `<!-- nuon-app-branch-preview:{app_branch_id} -->`. It is already
effectively per-PR, because the comment lives on the PR.

## UI

Two screens behind `/{orgId}/apps/{appId}/branches/{branchId}/pull-requests/{number}`:

1. **Configure** — where the first comment lands you. PR header, preview mode radio
   (`build-only` / `plan-only` / `apply`), install picker over the candidate list,
   GitHub toggles. Saves a `PreviewOverride`.
2. **Detail** — every later visit. The resolved config as a card (mode, install, and
   whether it came from the branch default or a PR override), over a table of every
   preview run on the PR: status, commit (short SHA, message, author), mode, install,
   trigger, and duration.

Proof of concept, with fixtures and no API calls:
`services/dashboard-ui/client/components/playground/pull-request/`, Ladle story
`Playground/PullRequest/Scope`.

## Out of scope

- Ephemeral per-PR installs, created on open and torn down on close.
- `issue_comment` webhook subscription for `/nuon plan`-style PR commands.
- Cross-app PR fan-out — one GitHub PR touching several apps. Deliberately deferred by
  the `(app_branch, pr_number)` identity choice.

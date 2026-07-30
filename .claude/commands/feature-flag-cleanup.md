---
name: feature-flag-cleanup
description: Find org feature flags defaulted to true and stable in prod, remove approved ones plus all gating code (ctl-api, dashboard-ui, CLI), and emit the admin bulk-enable request body
---

You are cleaning up stale org feature flags. A flag is stale when its default has been `true` long enough that the
gating code is dead weight. Cleanup means: remove the flag from the org model, remove every piece of gating code
across the monorepo, and produce the admin API request body needed to flip the flag on for existing orgs BEFORE the
removal deploys.

## Background: how the flag system works

- **Source of truth**: `services/ctl-api/internal/app/org.go`
  - `OrgFeature` string constants (e.g. `OrgFeatureRunbookStudio OrgFeature = "runbook-studio"`)
  - `GetFeatures()` — the active flag list (chronological, oldest first)
  - `GetFeatureDescriptions()` — flag → description map
  - `Org.BeforeCreate` — the `defaultFeatures` map where defaults are set; this is where a flag gets "defaulted to true"
  - `Org.AfterQuery` — auto-prunes keys not in `GetFeatures()` from `org.Features` at read time, so **no DB
    migration is needed** when a flag is removed
- **Gating patterns**:
  - ctl-api (Go): `org.Features[string(app.OrgFeatureXxx)]` (sometimes on a preloaded `cmp.Org`, `install.Org`, etc.)
  - dashboard-ui (TS): `org?.features?.['flag-name']` or dot access for single-word flags (`org?.features?.runbooks`)
  - CLI (`bins/cli`): no gating today, but always sweep it — search by the flag string literal
- **Flag-specific backend logic** can also live in `services/ctl-api/internal/pkg/features/` (e.g. the
  `control-plane-builds` / `org-runner` mutual-exclusion sync in `update.go`, sandbox logic in `sandbox.go`) and in
  config-driven overrides inside `Org.BeforeCreate` (e.g. `cfg.ControlPlaneBuildsDefaultEnabled`)
- **Admin bulk endpoint**: `PATCH /v1/orgs/admin-toggle-feature` (handler:
  `services/ctl-api/internal/app/orgs/service/admin_toggle_orgs_feature.go`) patches the given flags onto **every
  org** via jsonb merge. Body: `{"features": {"flag-name": true}}`. There is also a per-org variant at
  `PATCH /v1/orgs/{org_id}/admin-features`.
- **Critical ordering constraint**: both admin endpoints validate flag names against `GetFeatures()`. Once the
  removal deploys, the flag name is rejected as `invalid feature`. The PATCH must be applied to prod **before** the
  cleanup commit ships.

## Step 1: Identify candidates

1. Read the `defaultFeatures` map in `Org.BeforeCreate` in `org.go`. Collect every flag currently defaulted to
   `true`.
2. Drop any flag listed in the **Keep-list** at the bottom of this file.
3. Find the latest release tag and its date (tags mark promotion to prod):
   ```bash
   git tag --sort=-creatordate | head -1
   git log -1 --format=%cI <tag>
   ```
4. For each remaining flag, find when its default flipped to `true` (the map is gofmt-aligned, so match with a
   regex, not an exact string):
   ```bash
   git log --reverse -G 'OrgFeatureXxx:\s+true' --format='%h %cI %s' -- services/ctl-api/internal/app/org.go | head -1
   ```
   If the flag was `true` from the moment it was introduced, use the commit that added the constant instead.
5. A flag is a **candidate** when its flip-to-true commit date is at least 14 days older than the latest tag's
   date — i.e. the true default has been in prod for 2+ weeks.
6. For each candidate, count gating references so the user can judge blast radius:
   ```bash
   grep -rn "OrgFeatureXxx\|['\"]flag-name['\"]" --exclude-dir=node_modules --exclude-dir=graveyard --exclude-dir=.git .
   ```

## Step 2: Get approval

Present the candidates as a table: flag name, date the default flipped to true, days in prod ahead of the latest
tag, and reference counts split by area (ctl-api / dashboard-ui / other). Then ask the user, **in plain prose in a
single message** (never AskUserQuestion), for a yes/no on each flag.

For every flag the user declines, offer to add it to the Keep-list in this file so it stops appearing on future
runs. Only proceed with explicitly approved flags.

## Step 3: Remove approved flags and their gating code

For each approved flag, in this order:

1. **`org.go`**: remove the `OrgFeature` constant, its `GetFeatures()` entry, its `GetFeatureDescriptions()` entry,
   its `defaultFeatures` entry, and any flag-specific logic in `BeforeCreate`.
2. **Repo-wide sweep**: grep for both the Go constant name and the kebab-case string literal (the grep from step
   1.6). Cover ctl-api, dashboard-ui (`client/` SPA and legacy `src/`), `bins/cli`, `bins/runner`, docs, and seed
   configs. Skip `node_modules/`, `graveyard/`, and generated files (`sdks/`, `nuon-oapi` type defs — flags are
   plain JSONB map keys and don't appear in generated types).
3. **Collapse gating**: at every call site, the enabled path becomes unconditional and the disabled path is
   deleted. Handle negated checks (`if !org.Features[...]` → delete the block). Then chase the fallout: delete
   now-unreachable functions, components, routes, error helpers, and tests that only exercised the disabled path;
   update tests that toggled the flag. In dashboard-ui, watch for derived booleans (`const hasXxx = !!org?.features?.[...]`)
   that thread through props — inline `true` and simplify until the boolean disappears.
4. **Check `internal/pkg/features/`** for logic naming the flag (validation special cases, sync rules) and remove it.
5. **Verify**:
   ```bash
   gofmt -w ./services/ctl-api/... && goimports -w ./services/ctl-api/...
   go build ./services/ctl-api/... ./bins/...
   go test ./services/ctl-api/internal/app/orgs/... ./services/ctl-api/internal/pkg/features/...
   ```
   For dashboard-ui changes, lint only the changed files with
   `bunx oxlint -c client/.oxlintrc.json <files>` — do not run full type checking.
6. Re-run the grep from step 1.6 and confirm zero hits for the flag outside this skill file.

## Step 4: Generate the admin bulk-enable request body

Output the request the user runs against prod **before** merging/deploying the removal, so existing orgs that still
have the flag `false` get flipped on and the code removal becomes a behavioral no-op:

```
PATCH /v1/orgs/admin-toggle-feature
{
  "features": {
    "flag-a": true,
    "flag-b": true
  }
}
```

Include every removed flag in one body. Remind the user explicitly: this must be applied before the cleanup
deploys, because after removal the endpoint rejects the flag names as invalid.

## Keep-list

Flags that must never be proposed for cleanup (permanent platform flags, or ones the user has declined). Add
declined flags here when the user agrees.

<!-- keep-list:start -->
<!-- keep-list:end -->

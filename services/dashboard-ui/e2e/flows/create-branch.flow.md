# Flow: create-branch

Create an app branch (with a git/VCS config) from the app's branches page, then edit
the created branch's name. App-scoped; requires an app and a VCS connection in the org.

## Setup
- env: E2E_ORG_ID (required)
- prereq: org has at least one **VCS connection** (GitHub) — the create form's repository
  and git-branch selects are populated from it. Without a connection the modal shows the
  "No VCS connections found" banner and the "Create branch" button is disabled.
- prereq: org has at least one **app** (`appConfig` fixture holds the app *name*, e.g.
  "httpbin"). The fixtures expose `orgId`, `appConfig` (app name), and `installIds` — but
  **not** an appId, so the flow must reach the app via the Apps list (see first steps).
- start: /:orgId/apps

## Steps

### Open the apps list
- action: goto | /:orgId/apps
- expect: visible | heading "Apps"

### Open the app
- action: click | link "{appConfig}"
- expect: url | /apps/
- note: the App name cell in AppsTable renders a `Link` whose text is the app name
  (`getByRole('link', { name: appConfig })`). This lands on `/:orgId/apps/:appId`, so the
  appId becomes available in the URL for the remaining steps.

### Go to the branches page
- action: goto | {current app url}/branches
- expect: visible | heading "Branches"
- note: prefer deriving `:appId` from the app-detail URL captured above and navigating
  directly to `/:orgId/apps/:appId/branches` rather than hunting for a "Branches" nav link
  — the nav differs between the new-app-IA and legacy layouts. The Branches page always
  renders a "Branches" heading and a "Create branch" button.

### Open the create-branch modal
- action: click | button "Create branch"
- expect: visible | heading "Create app branch"

### Fill the branch name
- action: fill | input "Branch name" | e2e-branch
- note: field id `branch-name`, label "Branch name", placeholder "production".

### Select the repository
- action: select | select "Repository" | {first repo option}
- note: the "Repository" select (id `repo`, label "Repository") is populated from the
  chosen VCS connection. It is **searchable** and its option labels are the repo *name*
  (not the full owner/name). The value written to the form is `repo.full_name`. Because the
  concrete repo list depends on the test org's connection, the spec must pick the first
  available real option at runtime (open the select, choose the first non-empty option)
  rather than hardcoding a repo name. If the org has >1 VCS connection, a "VCS connection"
  select (id `vcs-connection`) also appears above it — leave the default (first connection)
  selected.

### Select the git branch
- action: select | select "Git branch" | {first branch option}
- note: id `git-branch`, label "Git branch". This field is polymorphic: when branches load
  it is a searchable `Select` of branch names; when none load (or the repo has none / the
  browse call errors) it falls back to a text `Input` with placeholder "main". The spec must
  handle both — if a select is present pick the first option, otherwise fill "main".

### Set the directory
- action: fill | input "Directory" | .
- note: id `directory`, label "Directory", prefilled with "." (or an autofilled value from
  an existing branch's config). Leaving "." is valid. Path filter (id `path-filter`,
  "Path filter (optional)") is optional and can be skipped.

### Submit
- action: click | button "Create branch"
- expect: visible | text "Branch created"
- note: on success a "Branch created" toast fires and the app redirects to the branch
  detail page `/:orgId/apps/:appId/branches/:branchId`. The `:branchId` is only in the URL
  — capture it here for phase 2. Also assert `url | /branches/` to confirm the redirect.

## Phase 2 — edit the created branch name (depends on the branch created above)

### Return to the branches list
- action: goto | /:orgId/apps/:appId/branches
- expect: visible | heading "Branches"
- note: the create flow leaves us on the branch *detail* page, where the edit affordance
  differs by IA mode (new-app-IA hides the header "Manage" dropdown and exposes edit only
  via the Settings side-panel; legacy renders a different overview). The **branches list**
  is consistent across both modes: every row/card renders a `BranchManagementDropdown` with
  an "Edit branch" menu item, so editing from the list is the reliable path.

### Open the created branch's management menu
- action: click | {row "e2e-branch"} → dots menu button
- expect: visible | button "Edit branch"
- note: SELECTOR UNCERTAINTY — the management dropdown trigger is a ghost icon button with
  **empty** text (DotsThreeVertical icon), so it can't be located by role+name. Scope to the
  branch's row first (`getByRole('row', { name: /e2e-branch/ })`) and click the last
  button in that row (the `…` menu), then assert the "Edit branch" menu item is visible.
  There is no testid on the dropdown today — if this proves flaky, add a `data-testid` to
  the `BranchManagementDropdown` trigger and locate by testid instead.

### Open the edit modal
- action: click | button "Edit branch"
- expect: visible | heading "Edit branch"

### Change the name
- action: fill | input "Branch name" | e2e-branch-renamed
- note: id `branch-name`, label "Branch name". The "Connect to git repository" checkbox and
  VCS fields are prefilled from the branch's current config — leave them as-is; renaming only
  requires changing the name field.

### Save
- action: click | button "Save changes"
- expect: visible | text "Branch updated"
- note: on success a "Branch updated" toast fires and the modal closes. Optionally assert the
  renamed branch is visible in the list (`visible | text "e2e-branch-renamed"`).

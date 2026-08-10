# Flow: run-adhoc-action

Core product flow. Run an **ad-hoc** action on an install — a free-form command/script
the user types, distinct from a predefined action (see [run-action](./run-action.flow.md),
which triggers the packaged `healthcheck` action via a confirm dialog). This flow opens the
"Run adhoc action" form on the Actions page, fills the minimal required input (a single
shell command), submits, and asserts redirect to the created workflow. Tier 2: assert the
redirect only — the ad-hoc action workflow needs a real runner and is never awaited.

## Setup
- fixtures: orgId, installIds
- install: installIds[0] (non-destructive, additive)
- skip-if: no seed install available (installIds[0] is undefined)
- start: /:orgId/installs/:installId/actions

## Steps

### open the actions page
- action: goto | /:orgId/installs/:installId/actions
- expect: title | /^Actions \|/

### open the adhoc action form
- action: click | button "Run adhoc action"
- expect: visible | dialog heading "Run adhoc action"

### dismiss any resumed draft
- note: the form persists a draft per-install (`adhoc-action-draft:<installId>`); if a
  previous run left a draft, a "Resume draft" modal appears first. Start fresh so the form is empty.
- action: if visible dialog heading "Resume draft" → click | button "Start fresh"
- skip-if: no "Resume draft" modal appears

### fill the command
- note: input mode defaults to "Single command"; the "Command *" field is the only required
  input. Timeout defaults to 300 and the execution role is optional.
- action: fill | dialog input "Command" | echo 'hello from e2e'

### submit
- action: click | dialog button "Run action"
- expect: url | /workflows/

## Notes
- Entry point: the "Run adhoc action" button on the Actions page (`views/install/Actions.tsx`).
  The same button also appears on the Workflows page and in the install management dropdown —
  the Actions page is the canonical, stable location.
- This is a DIFFERENT flow from run-action: run-action clicks a predefined action link then
  confirms in a "Run action <name>" dialog; this flow opens a free-form modal titled
  "Run adhoc action" and submits user-typed command input.
- On success the container navigates to `/:orgId/installs/:installId/workflows/:workflowId`
  (falls back to `/workflows` if the response has no `workflow_id`). Assert the `/workflows/`
  URL, not the toast — the "Adhoc action started" toast races with navigation.
- The submit button label is "Run action" for a fresh form ("Rerun action" only when
  `initialValues` are prefilled, which does not happen from the Actions page entry point).
- Do NOT touch the "Bash script" tab, timeout, environment variables, or role — they are
  optional and outside the minimal happy path.

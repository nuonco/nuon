# Flow: trigger-branch-run

Trigger a run for an app branch. Unlike other Tier-2 flows this does NOT redirect —
the modal closes and a "Run triggered" toast fires while the branch page refreshes,
so the toast is the durable signal here.

## Setup
- fixtures: orgId
- feature: app-branches-ui (enabled on the e2e org in global-setup)
- isolation: `nuon apps sync` does NOT create a branch for httpbin, so the spec seeds one
  via the API (helpers.createTriggerableBranch): it creates its own throwaway install, a
  branch, and a branch config with one install group (an install can only belong to one
  branch, so a dedicated install avoids collisions). The install group is what makes the
  branch's deployment plan present, which enables the "Trigger run" button.
- start: /:orgId/apps/:appId/branches/:branchId

## Steps

### open the seeded branch
- action: goto | /:orgId/apps/:appId/branches/:branchId

### open the trigger modal
- action: click | button "Trigger run" (must be enabled — requires a deployment plan)
- expect: visible | dialog button "Trigger run"

### trigger
- action: click | dialog button "Trigger run"
- expect: visible | text "Run triggered"

## Notes
- The page button, modal heading, and modal confirm all read "Trigger run" — assert/act on
  the dialog's confirm button, not the heading text (avoids strict-mode collisions).

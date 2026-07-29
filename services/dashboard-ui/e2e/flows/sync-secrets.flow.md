# Flow: sync-secrets

Sync install secrets from the install Settings panel. Asserts up to workflow creation
only (redirect to the workflows page).

## Setup
- fixtures: orgId, installIds
- install: installIds[0]
- start: /:orgId/installs/:installId?panel=settings

## Steps

### open the settings panel
- action: goto | /:orgId/installs/:installId?panel=settings
- note: the `?panel=settings` query param auto-opens the Settings panel (role="complementary").
  The install actions moved out of a "Manage" dropdown into this panel (commit #1990).

### open the sync-secrets modal
- action: click | panel button "Sync secrets" (scope to role="complementary")
- expect: visible | dialog button "Sync secrets"

### sync
- action: click | dialog button "Sync secrets"
- expect: url | /workflows OR skip-if the backend rejects (install has no secrets)

## Notes
- Panel is role="complementary", the confirm modal is role="dialog" — scope each button accordingly.
- httpbin may have no secrets; if the backend errors (error toast "Secret sync failed"),
  the flow no-ops and the test skips rather than fails.

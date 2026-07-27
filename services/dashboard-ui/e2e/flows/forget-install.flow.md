# Flow: Forget install

Forgets an install from the Manage dropdown (type-to-confirm the install name) and verifies the redirect back to the installs list. Forget removes the install record but does NOT deprovision cloud resources, so it completes deterministically in a sandbox.

## Setup
- requires: a SECOND seed install (installIds[1]) — skip otherwise
- isolation: uses installIds[1] (never installIds[0]) so it does not corrupt the installs that the other install-scoped specs run against in parallel
- start: /:orgId/installs/{installId1}

## Steps

### Navigate to the second install's page
- action: goto | /:orgId/installs/{installId1}
- action: wait | domcontentloaded
- expect: title | /^Overview \|/

### Open Forget from the Manage dropdown
- action: click | button "Manage"
- action: click | button "Forget install"
- expect: visible | text "Forget"

### Type-to-confirm the install name and forget
- action: fill | input "install name" | <the install name, read from the modal heading>
- action: click | button "Forget install" last

### Redirected to the installs list (the success signal; the toast races with navigation)
- expect: url | /installs

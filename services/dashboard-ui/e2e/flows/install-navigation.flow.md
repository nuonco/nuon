# Flow: Install navigation

Verifies each install detail sub-page is reachable and renders without the error boundary fallback. Each page sets a distinct browser tab title via `PageTitle`, used as the positive "rendered correctly" assertion.

## Setup
- requires: at least one seed install (installIds[0]) — skip otherwise
- start: /:orgId/installs/{installId}

## Steps

### Overview
- action: goto | /:orgId/installs/{installId}
- action: wait | domcontentloaded
- expect: title | /^Overview \|/
- expect: not-visible | text "Something went wrong."

### Current inputs
- action: goto | /:orgId/installs/{installId}/inputs
- action: wait | domcontentloaded
- expect: title | /^Current inputs \|/
- expect: visible | text "The current input values for this install."
- expect: not-visible | text "Something went wrong."

### View state
- action: goto | /:orgId/installs/{installId}/state
- action: wait | domcontentloaded
- expect: title | /^State \|/
- expect: not-visible | text "Something went wrong."

### Components
- action: goto | /:orgId/installs/{installId}/components
- action: wait | domcontentloaded
- expect: title | /^Components \|/
- expect: not-visible | text "Something went wrong."

### Actions
- action: goto | /:orgId/installs/{installId}/actions
- action: wait | domcontentloaded
- expect: title | /^Actions \|/
- expect: not-visible | text "Something went wrong."

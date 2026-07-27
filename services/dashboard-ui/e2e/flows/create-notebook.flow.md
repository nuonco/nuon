# Flow: Create notebook

Creates a notebook on an install and verifies the redirect to the notebook detail page. Pure API/DB flow.

## Setup
- requires: at least one seed install (installIds[0]) — skip otherwise
- note: global-setup enables the `notebooks` org feature on the created test org
- start: /:orgId/installs/{installId}/notebooks

## Steps

### Navigate to the notebooks page
- action: goto | /:orgId/installs/{installId}/notebooks
- action: wait | domcontentloaded
- expect: title | /^Notebooks \|/

### Open the create notebook modal
- action: click | button "Create notebook" first
- expect: visible | text "Create notebook"

### Fill the name and submit
- action: fill | input "e.g. Debug pods" | e2e-notebook-{timestamp}
- action: click | button "Create notebook" last
- expect: visible | text "Notebook created"

### Redirected to the notebook detail page
- expect: url | /notebooks/

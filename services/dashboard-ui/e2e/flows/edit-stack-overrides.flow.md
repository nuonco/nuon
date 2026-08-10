# Flow: edit-stack-overrides

Edit an install's stack overrides from the install Stacks page: open the "Edit
stack overrides" modal, change the VPC nested template URL, save, and assert the
"Stack overrides updated" success toast.

## Setup
- env: E2E_ORG_ID (required)
- fixture: installIds — requires an existing seed install; use `installIds[0]`
- skip: `test.skip(!installIds[0], "No seed install available")` (same guard as run-action.spec.ts)
- start: /:orgId/installs/:installId/stacks (installId = installIds[0])
- MUTATES the seed install: writes a value into the install config's
  `vpc_nested_template_url` override. Keep the change minimal — set a single
  placeholder S3 URL. Not auto-reverted; documented as a benign, idempotent write.
- NOTE: if the install is managed by config (`metadata.managed_by ==
  'nuon/cli/install-config'`) the trigger renders as a disabled button wrapped in
  a tooltip and the modal never opens. Seed installs are not config-managed, but
  the skip guard only covers "no install", not "config-managed install".

## Steps

### Open the install Stacks page
- action: goto | /:orgId/installs/:installId/stacks
- action: wait | domcontentloaded
- expect: title | "Stacks |"
- expect: visible | text "Install stacks"

### Open the edit stack overrides modal
- action: click | button "Edit stack overrides"
- expect: visible | heading "Edit stack overrides"

### Change the VPC nested template URL
- action: fill | first textbox in dialog | https://s3.amazonaws.com/nuon-e2e/vpc-template.yaml
- note: the VPC and Runner URL fields are bare `<input>` elements with no
  associated `<label>`, so `getByLabel` will NOT work. Scope to the dialog and
  target the first `textbox` role: `dialog.getByRole('textbox').first()`. The
  second textbox is the Runner URL. (Custom-stack Name/Template inputs only
  appear after clicking "Add stack".)

### Save the overrides
- action: click | button "Save overrides"
- expect: visible | text "Stack overrides updated"
- expect: not-visible | heading "Edit stack overrides"

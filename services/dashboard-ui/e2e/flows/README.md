# Flow Spec Format

Structured markdown format for describing E2E test flows. These serve as source-of-truth documentation — when flows change, update the markdown and ask Claude Code to regenerate the Playwright spec.

## Flows

Org-level:

| Flow | Spec | Covers |
|------|------|--------|
| [smoke](./smoke.flow.md) | `specs/smoke.spec.ts` | Auth cookie injection and basic page render |
| [navigation](./navigation.flow.md) | `specs/navigation.spec.ts` | Every org-level nav page is reachable and titled |
| [webhooks-crud](./webhooks-crud.flow.md) | `specs/webhooks.spec.ts` | Create, edit, and delete a webhook |
| [create-api-token](./create-api-token.flow.md) | `specs/api-tokens.spec.ts` | Create an API token, confirm the reveal modal, then delete it |

Install-level:

| Flow | Spec | Covers |
|------|------|--------|
| [create-install](./create-install.flow.md) | `specs/create-install.spec.ts` | Create an install from an app and redirect to its provision workflow |
| [install-navigation](./install-navigation.flow.md) | `specs/install-navigation.spec.ts` | Every install sub-page renders without the error boundary |
| [install-quick-management](./install-quick-management.flow.md) | `specs/install-quick-management.spec.ts` | Install row dropdown actions (edit inputs, current inputs, view state) |
| [run-action](./run-action.flow.md) | `specs/run-action.spec.ts` | Trigger a manual action and redirect to its workflow |
| [run-runbook](./run-runbook.flow.md) | `specs/run-runbook.spec.ts` | Run a runbook and redirect to its workflow |

Tier 2 — workflow-triggering (assert redirect to the workflows page only, never await the workflow):

| Flow | Spec | Covers |
|------|------|--------|
| [deploy-component](./deploy-component.flow.md) | `specs/deploy-component.spec.ts` | Deploy a component build and redirect to its workflow |
| [drift-scan-component](./drift-scan-component.flow.md) | `specs/drift-scan-component.spec.ts` | Drift scan a component and redirect to its workflow |
| [sync-secrets](./sync-secrets.flow.md) | `specs/sync-secrets.spec.ts` | Sync install secrets and redirect to its workflow |
| [trigger-branch-run](./trigger-branch-run.flow.md) | `specs/trigger-branch-run.spec.ts` | Trigger an app branch run (asserts the "Run triggered" toast — no redirect) |
| [teardown-component](./teardown-component.flow.md) | `specs/teardown-component.spec.ts` | Teardown a component on a throwaway install |
| [reprovision-install](./reprovision-install.flow.md) | `specs/reprovision-install.spec.ts` | Reprovision a throwaway install |
| [deprovision-install](./deprovision-install.flow.md) | `specs/deprovision-install.spec.ts` | Deprovision a throwaway install |
| [sandbox-management](./sandbox-management.flow.md) | `specs/sandbox-management.spec.ts` | Drift scan, reprovision, and deprovision the sandbox on a throwaway install |

Destructive Tier-2 specs (teardown, reprovision, deprovision, sandbox) create their own throwaway
install via the API (`helpers.ts`) so they don't corrupt the shared seed installs; global teardown
drops the org.

Some specs have no flow doc yet: `specs/apps.spec.ts` and `specs/install-labels.spec.ts`.

## Format

```markdown
# Flow: <name>

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/installs

## Steps

### <step name>
- action: goto | /:orgId/apps
- expect: visible | heading "Apps"

### <step name>
- action: click | button "Create install"
- expect: visible | text "Select an app"
```

## Action types

| Action | Syntax | Description |
|--------|--------|-------------|
| `goto` | `goto \| /path` | Navigate to URL |
| `click` | `click \| button "Label"` | Click element by role + name |
| `fill` | `fill \| input "Label" \| value` | Fill input by label |
| `select` | `select \| select "Label" \| option` | Select dropdown option |
| `wait` | `wait \| domcontentloaded` | Wait for condition (never `networkidle` — the SPA polls continuously, so the network never goes idle) |

## Assertion types

| Assertion | Syntax | Description |
|-----------|--------|-------------|
| `visible` | `visible \| heading "Text"` | Element is visible |
| `not-visible` | `not-visible \| text "Error"` | Element is not visible |
| `title` | `title \| "Page Title"` | Page title matches |
| `url` | `url \| /apps` | URL contains path |
| `count` | `count \| row \| 3` | Element count matches |

## Locator types

Used in actions and assertions after the `|` separator:

- `heading "Text"` — `getByRole('heading', { name: 'Text' })`
- `button "Text"` — `getByRole('button', { name: 'Text' })`
- `link "Text"` — `getByRole('link', { name: 'Text' })`
- `text "Text"` — `getByText('Text')`
- `input "Label"` — `getByLabel('Label')`
- `select "Label"` — `getByLabel('Label')`
- `testid "id"` — `getByTestId('id')`
- `row` — `getByRole('row')`
- `.class-name` — `locator('.class-name')`

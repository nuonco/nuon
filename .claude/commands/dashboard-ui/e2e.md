---
name: dashboard-ui:e2e
description: Create a Playwright E2E test from a user-described flow
---

This skill creates a Playwright E2E spec by first writing a flow markdown doc, getting user approval, then generating the spec and running it.

## Step 1: Gather the flow

Ask the user (use AskUserQuestion):

1. **What flow do you want to test?** — Describe the user journey in plain language (e.g. "navigate to installs, open create install modal, verify app selection appears")

## Step 2: Write the flow markdown

Create `services/dashboard-ui/e2e/flows/<name>.flow.md` using the format defined in `e2e/flows/README.md`.

Before writing, read `services/dashboard-ui/e2e/flows/README.md` to understand the action/assertion syntax.

If the flow references specific UI elements, read the relevant view or component source files to find the correct selectors (text content, roles, class names, test IDs).

Show the user the flow file you created and ask: **"Does this flow look right? I'll generate the Playwright spec from it next."**

Wait for confirmation before proceeding.

## Step 3: Generate the spec

Create `services/dashboard-ui/e2e/specs/<name>.spec.ts` that implements the flow.

Rules:
- Import `{ test, expect }` from `../fixtures` — never from `@playwright/test`
- Use the `orgId` fixture for org-scoped navigation
- Use `page.waitForLoadState('networkidle')` after navigation
- Prefer Playwright's role-based locators (`getByRole`, `getByText`, `getByLabel`) over CSS selectors
- Keep tests focused — one `test()` per logical step or group of related assertions
- Use `test.describe` if the flow has multiple distinct phases

## Step 4: Run the test

Run the new spec:

```bash
cd services/dashboard-ui && E2E_EMAIL=${E2E_EMAIL} E2E_ORG_ID=${E2E_ORG_ID} npx playwright test -c e2e/playwright.config.ts e2e/specs/<name>.spec.ts
```

If the test fails, read the error output and fix the spec. Do not modify the flow markdown unless the user asks — the flow is the source of truth.

## Anti-Patterns

- **Do not** generate the spec without writing the flow markdown first — the flow is the source of truth
- **Do not** generate the spec without user approval of the flow — the checkpoint prevents wasted iteration
- **Do not** import from `@playwright/test` directly — always use `../fixtures`
- **Do not** use CSS selectors when a role-based locator works — prefer `getByRole('button', { name: 'Create' })` over `locator('.btn-create')`
- **Do not** modify the flow markdown to fix a failing spec — fix the spec to match the flow, or ask the user if the flow should change

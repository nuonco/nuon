---
name: dashboard-ui:add-tracking
description: Use when adding PostHog analytics tracking to a user action (mutation) in the dashboard-ui
---

Add a `trackEvent` call to a mutation so a user action is captured in PostHog. Analytics is proxied
through the BFF and is off unless the org runs on Nuon-owned cloud (never in BYOC), so tracking is
always best-effort — `trackEvent` is a silent no-op when PostHog isn't initialized. You never need
to guard call sites.

Canonical source: `client/lib/posthog-analytics.tsx`. Reference call site:
`client/components/install-components/management/DeployComponent/DeployComponentContainer.tsx`.

## Steps

1. **Find the mutation** for the action in its container (`useMutation` with `onSuccess` / `onError`).
   Tracking lives in the container next to the mutation, never in the presentational component.

2. **Import and pull `user`:**
   ```typescript
   import { useAuth } from '@/hooks/use-auth'
   import { trackEvent } from '@/lib/posthog-analytics'
   // ...
   const { user } = useAuth()
   ```

3. **Fire on BOTH success and failure** with the same event name and `status: 'ok' | 'error'`. If the
   mutation has no `onError` yet, add one just for tracking (errors still surface via toast / form
   banner elsewhere — don't move that logic):
   ```typescript
   onSuccess: (result) => {
     trackEvent({
       event: 'install_create',
       status: 'ok',
       user,
       props: { appId: app.id, installId: result.data.id },
     })
     // ...existing toast + invalidateQueries + navigation
   },
   onError: (err: any) => {
     trackEvent({
       event: 'install_create',
       status: 'error',
       user,
       props: { appId: app.id, err: err?.error },
     })
   },
   ```

4. **Event name = snake_case**, `resource_verb`, singular vs plural mirroring the action (e.g.
   `component_deploy`, `components_deploy`, `runner_update`, `install_create`). Grep existing names
   first (`grep -rn "event: '" client/components/`) and match the established scheme; don't invent a
   new casing or word order.

5. **Props: only what super properties don't already carry.** `org_id`, `app_id`, and `install_id`
   are auto-attached to every event as super properties **when those contexts are active** (org
   always, app/install only inside their providers). So:
   - Pass `appId` / `installId` explicitly **only** when the action runs outside that context (e.g.
     `install_create` fires from the apps or installs list, where no InstallProvider exists yet).
   - Otherwise pass just the action-specific ids (`componentId`, `branchId`, `runnerId`, …).
   - On error, add `err: err?.error`.
   - **Pass camelCase keys** — the adaptor snake_cases them automatically (`branchName` →
     `branch_name`). Do NOT hand-snake_case prop keys; only the event **name** is snake_case.

6. **Keep the `trackEvent` shape exactly** `{ event, status, user, props }`. Do not import `posthog-js`
   directly or call `posthog.capture` at a call site — always go through `trackEvent`.

## Anti-Patterns

- **Do not** guard the call site (`if (posthogEnabled) …`) — `trackEvent` no-ops itself when off.
- **Do not** track only success — fire on `onError` too so failure rate is measurable.
- **Do not** hand-snake_case prop keys or duplicate `org_id` / `app_id` / `install_id` when the
  context provider is already active (the super property covers it).
- **Do not** import `posthog-js` or reference the write key in a component — only
  `posthog-analytics.tsx` touches PostHog directly.
- **Do not** put `trackEvent` in a presentational component — it belongs in the container with the
  mutation.

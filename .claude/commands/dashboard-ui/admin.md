---
name: dashboard-ui:admin
description: Add a new admin action to the dashboard-ui admin panel
---

You are adding a new admin action to the dashboard-ui admin panel.

## Step 1: Gather requirements

Ask the user these questions (use AskUserQuestion with all 4 at once):

1. **Which section?** — Org, App, or Install
2. **What does the action do?** — Free-text description (e.g. "force sync all components")
3. **Which API endpoint?** — Method and path (e.g. `POST installs/:installId/admin-force-sync`)
4. **Variant?** — `default`, `warning`, or `danger`

## Step 2: Add the API function

Edit `client/lib/ctl-api/admin/index.ts`.

### Types and helpers already in the file — never duplicate these:

```typescript
type AdminBase = { adminApiUrl: string }
type AdminMutation = AdminBase & { adminEmail: string }
const adminHeaders = (adminEmail: string) => ({ 'X-Nuon-Admin-Email': adminEmail })
```

### Choosing the right signature

**Most actions** use the admin API (requires `baseUrl` + `X-Nuon-Admin-Email`):

```typescript
export const adminMyAction = ({ installId, adminApiUrl, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/admin-my-action` })
```

**Rare: public API actions** use `orgId` instead of `baseUrl`/`adminEmail` (like `adminTeardownInstallComponents`):

```typescript
export const adminMyAction = ({ installId, orgId }: { installId: string; orgId: string }) =>
  api<void>({ method: 'POST', body: {}, orgId, path: `installs/${installId}/my-action` })
```

Use the entity ID that matches the section:
- **Org section** → `orgId`
- **App section** → `appId`
- **Install section** → `installId` (may also need `runnerId` for runner-specific actions)

GET actions use `AdminBase` (no email needed). POST/PATCH/DELETE actions use `AdminMutation`.

## Step 3: Wire into the section component

Edit the correct section file:

| Section | File | Available IDs |
|---------|------|---------------|
| Org | `client/components/admin/sections/AdminOrgSection.tsx` | `orgId` |
| App | `client/components/admin/sections/AdminAppSection.tsx` | `orgId`, `appId` |
| Install | `client/components/admin/sections/AdminInstallSection.tsx` | `orgId`, `installId`, `runnerId` |

### Pattern

1. Import the new function from `@/lib`
2. Add an `<AdminActionCard>` inside an existing `<AdminActionGroup>`, or create a new group if the action doesn't fit any existing group
3. Pass the action as a closure: `action={() => adminMyAction({ appId, adminApiUrl, adminEmail })}`

### Confirmation rules by variant

| Variant | `requiresConfirmation` | `requiresInput` | `inputText` |
|---------|----------------------|-----------------|-------------|
| `default` | optional | no | — |
| `warning` | **always** | no | — |
| `danger` | **always** | **always** | `"CONFIRM"` |

### Example (canonical — from AdminAppSection):

```tsx
<AdminActionCard
  title="Reprovision app"
  description="Reprovision current app infrastructure"
  action={() => adminReprovisionApp({ appId, adminApiUrl, adminEmail })}
  variant="warning"
  requiresConfirmation
  confirmationText="This will reprovision the app infrastructure. This may affect all installs of this app."
/>
```

### Example (danger variant):

```tsx
<AdminActionCard
  title="Force shutdown runner"
  description="Immediately terminate the runner process"
  action={() => adminForceRunnerShutdown({ runnerId, adminApiUrl, adminEmail })}
  variant="danger"
  requiresConfirmation
  requiresInput
  inputText="CONFIRM"
  confirmationText="This will forcefully shut down the runner. Any in-progress jobs will be lost."
/>
```

## Anti-patterns — do NOT do these

- **Never skip `requiresConfirmation` for `warning` or `danger` variants** — these are destructive or disruptive actions
- **Never duplicate `AdminBase`, `AdminMutation`, or `adminHeaders`** — they already exist in the admin index file
- **Never create new section files** — add to the existing Org/App/Install sections
- **Never import from `@/lib/ctl-api/admin/index` directly** — import from `@/lib` (barrel export)
- **Never use `baseUrl` + `adminEmail` when the endpoint is on the public API** — use `orgId` instead (see `adminTeardownInstallComponents` for the pattern)

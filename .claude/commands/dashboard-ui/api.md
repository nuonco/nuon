---
name: dashboard-ui:api
description: Use when creating a new API function in client/lib/ctl-api/
---

This skill enforces creating a lib/ctl-api API function with correct types and barrel export.

## Steps

1. Determine the resource domain (e.g., runners, installs, apps) and place the new file at `client/lib/ctl-api/<domain>/<function-name>.ts`.
2. Write the function using this pattern — named export, destructured params with inline types, returns `api<T>({...})`:
   ```typescript
   import { api } from '@/lib/api'
   import type { TMyResource } from '@/types'

   export const getMyResource = ({
     resourceId,
     orgId,
   }: {
     resourceId: string
     orgId: string
   }) =>
     api<TMyResource>({
       path: `my-resources/${resourceId}`,
       orgId,
     })
   ```
3. Import the `T` type from `@/types` (which re-exports from `ctl-api.types.ts`). Never import from `nuon-oapi-v3.d.ts` directly.
4. `api<T>()` returns `T` directly — not `{ data: T }`. Access fields as `result.id`, not `result.data.id`.
5. Add the new function to the domain's `index.ts` barrel: `export * from './get-my-resource'`

## Anti-Patterns

- **Do not** import types directly from `nuon-oapi-v3.d.ts` — import from `@/types`
- **Do not** skip the barrel export in `index.ts` — functions not in the barrel are not importable from `@/lib`
- **Do not** write `result.data.id` — `api<T>()` returns `T` directly, not wrapped in `{ data }`

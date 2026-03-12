---
name: dashboard-ui:api
description: Use when creating a new API function in client/lib/ctl-api/
---

This skill enforces creating a lib/ctl-api API function and its co-located unit test together, plus adding the barrel export.

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
5. If the endpoint is new and not yet in `test/mock-api-handlers.js`, run `npm run generate-api-mocks` to refresh the MSW handlers before writing the test.
6. Create `client/lib/ctl-api/<domain>/<function-name>.test.ts` in the same step as the function file:
   ```typescript
   import { badResponseCodes } from '@test/utils'
   import { describe, expect, test } from 'vitest'
   import { getMyResource } from './get-my-resource'

   describe('getMyResource should handle response status codes from GET my-resources/:id endpoint', () => {
     const orgId = 'test-org-id'
     const resourceId = 'test-resource-id'

     test('200 status', async () => {
       const result = await getMyResource({ orgId, resourceId })
       expect(result).toHaveProperty('id')
     })

     test.each(badResponseCodes)('%s status', async () => {
       await expect(getMyResource({ orgId, resourceId })).rejects.toMatchObject({
         error: expect.any(String),
         description: expect.any(String),
         user_error: expect.any(Boolean),
       })
     })
   })
   ```
7. Add the new function to the domain's `index.ts` barrel: `export * from './get-my-resource'`
8. Run `npm run test:spa -- --run client/lib/ctl-api/<domain>/` to verify both the 200-status and error-status tests pass.

## Anti-Patterns

- **Do not** create the function file without the co-located `.test.ts` — they are always created together
- **Do not** import types directly from `nuon-oapi-v3.d.ts` — import from `@/types`
- **Do not** skip the barrel export in `index.ts` — functions not in the barrel are not importable from `@/lib`
- **Do not** write `result.data.id` — `api<T>()` returns `T` directly, not wrapped in `{ data }`
- **Do not** hardcode `[[400], [401], [403], [404], [500]]` in tests — use `badResponseCodes` from `@test/utils`

---
name: dashboard-ui:pagination
description: Use when adding pagination to a list view in the dashboard-ui
---

This skill enforces URL-driven pagination using useSearchParams for offset, keepPreviousData to prevent flash, and the Table or Timeline pagination prop.

## Steps

1. Define a `LIMIT` constant at the top of the file:
   ```typescript
   const LIMIT = 20
   ```

2. Read the current offset from the URL — never from `useState`:
   ```typescript
   const [searchParams] = useSearchParams()
   const offset = Number(searchParams.get('offset') ?? 0)
   ```

3. Include `offset` in the `queryKey` so a URL change automatically triggers a new fetch:
   ```typescript
   queryKey: ['my-resources', org.id, offset]
   ```

4. Add `placeholderData: keepPreviousData` to the `useQuery` call:
   ```typescript
   const { data: result, isLoading } = useQuery({
     queryKey: ['my-resources', org.id, offset],
     queryFn: () => getMyResources({ orgId: org.id, offset, limit: LIMIT }),
     placeholderData: keepPreviousData,
   })
   ```

5. Pass the `pagination` prop to `Table` (or `Timeline`):
   ```typescript
   <Table
     columns={columns}
     data={result?.data ?? []}
     isLoading={isLoading}
     pagination={{ hasNext: result?.pagination?.hasNext ?? false, offset, limit: LIMIT }}
   />
   ```

6. Do NOT add `PaginationProvider` manually — `Table` and `Timeline` wrap it internally.

7. Do NOT write custom "Previous / Next" buttons — the `Pagination` component rendered by `Table`/`Timeline` handles URL navigation automatically.

Canonical source: `client/components/installs/InstallsTable.tsx`

## Anti-Patterns

- **Do not** store page offset in `useState` — it must live in the URL via `useSearchParams` so pagination is bookmarkable and the browser back button works
- **Do not** omit `keepPreviousData` — without it the list flashes to empty on every page change while the next fetch is in-flight
- **Do not** omit `offset` from the `queryKey` — without it, changing the URL does not trigger a new fetch
- **Do not** add `PaginationProvider` in the component — `Table` and `Timeline` already wrap it internally

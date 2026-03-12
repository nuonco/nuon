---
name: dashboard-ui-view
description: Use this skill when adding a new page or view to the dashboard-ui client/ SPA.
model: sonnet
color: blue
---

This skill enforces correct route registration, layout-aware provider usage, and a guarded useQuery pattern when adding a new view.

## Steps

1. Decide the URL and layout level. Install-level pages (under `:orgId/installs/:installId/`) go in `client/views/install/routes.tsx` as a child of `{ element: <InstallLayout />, children: [...] }`. Org-level pages go in `client/views/org/routes.tsx`.

2. Add the route entry BEFORE creating the view component file:
   ```tsx
   { path: ':orgId/installs/:installId/my-page', element: <MyPage /> }
   ```

3. Create the view file at `client/views/install/MyPage.tsx` (or `client/views/org/` for org-level).

4. Read context from provider hooks — never from props passed down from a parent:
   - `const { org } = useOrg()`
   - `const { install } = useInstall()` (install-level only)
   - `const { resourceId } = useParams()`

5. Fetch data with `useQuery`, always including an `enabled` guard:
   ```typescript
   const { data: resource } = useQuery({
     queryKey: ['my-resource', org?.id, resourceId],
     queryFn: () => getMyResource({ orgId: org.id, resourceId: resourceId! }),
     enabled: !!org?.id && !!resourceId,
   })
   ```

6. Do NOT add `SurfacesProvider` or `ToastProvider` inside the view — they are already provided by `InstallLayout`. Adding them again creates a nested context that breaks `useSurfaces()` lookups.

7. Wrap page content in `<PageSection isScrollable>` for consistent scroll behavior.

## Anti-Patterns

- **Do not** register an install-level route outside `InstallLayout.children` — the view will render without its providers
- **Do not** omit the `enabled` guard on `useQuery` — `org` and `install` are `undefined` on the first render before providers hydrate, causing "Cannot read properties of undefined"
- **Do not** add `SurfacesProvider` or `ToastProvider` in a child view — `InstallLayout` already provides them
- **Do not** call `useInstall()` outside a route that is a child of `InstallLayout` — the provider won't be present

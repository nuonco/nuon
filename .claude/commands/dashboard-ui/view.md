---
name: dashboard-ui:view
description: Use when adding a new page or view to the dashboard-ui client/ SPA
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

5. Import **container** components (the default export from component directories), not presentational components. The container handles data-fetching; the view just composes containers:
   ```typescript
   // ✅ Correct — imports the container via barrel
   import { MyComponent } from '@/components/domain/MyComponent'

   // ❌ Wrong — imports the presentational component directly
   import { MyComponent } from '@/components/domain/MyComponent/MyComponent'
   ```

6. Fetch data with `useQuery`, always including an `enabled` guard:
   ```typescript
   const { data: resource } = useQuery({
     queryKey: ['my-resource', org?.id, resourceId],
     queryFn: () => getMyResource({ orgId: org.id, resourceId: resourceId! }),
     enabled: !!org?.id && !!resourceId,
   })
   ```

7. Do NOT add `SurfacesProvider` or `ToastProvider` inside the view — they are already provided by `InstallLayout`. Adding them again creates a nested context that breaks `useSurfaces()` lookups.

8. Use the correct page structure based on the route level:

   **Org-level page** (has its own PageLayout):
   ```tsx
   export const MyPage = () => (
     <PageLayout>
       <PageTitle title="My page" />
       <PageHeader>
         <PageHeadingGroup title="My page" />
       </PageHeader>
       <PageContent>
         <PageSection>
           {/* content */}
         </PageSection>
       </PageContent>
     </PageLayout>
   )
   ```

   **Child page inside App/Install layout** (just content, no PageLayout):
   ```tsx
   export const MyChildPage = () => {
     const { install } = useInstall()
     return (
       <PageSection>
         <PageTitle segments={['My page', install?.name]} />
         {/* content */}
       </PageSection>
     )
   }
   ```

   **Detail page with flush header** (inside App/Install layout):
   ```tsx
   export const MyDetailPage = () => {
     const { install } = useInstall()
     return (
       <>
         <PageTitle segments={[resource?.name, install?.name]} />
         <PageSection flush>
           <MyHeader />
         </PageSection>
         <PageSection>
           {/* content */}
         </PageSection>
       </>
     )
   }
   ```

   Scrolling, BackToTop, and SubNav sticky are all handled automatically by PageLayout — do not add them manually.

## Page title (required, UXDR 018)

**Every routed view sets its own `document.title`. Layouts NEVER set it — the leaf view that renders the page owns its title.** (This is why tab pages under a shared layout each set their own; the layout does not.)

The provider appends `| Nuon`, so a view supplies at most two segments, **most specific first**:

`{specific} | {owning entity}` →  `"Deploy logs | acme-install"`, `"Components | acme-app"`

- **`{specific}`** — sentence-case page name for section pages (`'Components'`, `'API tokens'`); the entity's own name for detail pages (`resource?.name`); for a **tab page**, fold the parent context in: `'Deploy logs'`, `'Sandbox run plan'`, `` `${runbook?.name} steps` ``.
- **`{owning entity}`** — the install or app name from `useInstall()` / `useApp()`. **Org-level pages have NO owner segment** (the org name is never a title segment) — pass a single element: `<PageTitle title="Webhooks" />`.
- Unset segments are dropped automatically — pass `install?.name` (not a guarded string); it's omitted while loading, never rendered as `"undefined"`.

**Always `<PageTitle>`, rendered as the first element the view returns.** It's headless (returns `null`), so it only needs to be in the rendered output.

For views with **early returns** (loading/empty/error branches, common in tab panels), wrap the body in a **fragment** so the title renders on every branch — never bury `<PageTitle>` in only the happy-path return, or the title goes stale on the loading/empty branches:

```tsx
// Section / detail view
<PageTitle segments={['Components', install?.name]} />

// Tab panel with early returns — fragment so the title always renders
export const DeployPlanTab = () => {
  const { install } = useInstall()
  return (
    <>
      <PageTitle segments={['Deploy plan', install?.name]} />
      {isLoading ? <Skeleton /> : !plan ? <EmptyState … /> : <Plan … />}
    </>
  )
}
```

## Redirects

Use a `loader` with `redirect` from `react-router` — never `<Navigate>`:

```tsx
import { redirect, type RouteObject } from 'react-router'

// ✅ Correct
{ path: ':orgId/connections', loader: ({ params }) => redirect(`/${params.orgId}`) }

// ❌ Wrong
{ path: ':orgId/connections', element: <Navigate to=".." replace /> }
```

See `client/views/install/routes.tsx` for examples.

## Loading states

A page's loading state is the page itself with `loading` primitives inside — breadcrumbs, headings, and tabs render real immediately; each region renders its normal components in loading mode. No full-page spinners, no hand-built page-skeleton blocks. Use the primitive `loading` prop, `<Table isLoading>` / `Timeline` for collections, and `<Loading variant="large" />` only for unknown-shape content. Add `placeholderData: keepPreviousData` to list/detail `useQuery` sites so revisits skip the cold load (leave SSE-backed views alone — they already write the cache). See `DESIGN.md` §5 "Loading states".

## Anti-Patterns

- **Do not** register an install-level route outside `InstallLayout.children` — the view will render without its providers
- **Do not** omit the `enabled` guard on `useQuery` — `org` and `install` are `undefined` on the first render before providers hydrate, causing "Cannot read properties of undefined"
- **Do not** add `SurfacesProvider` or `ToastProvider` in a child view — `InstallLayout` already provides them
- **Do not** call `useInstall()` outside a route that is a child of `InstallLayout` — the provider won't be present
- **Do not** add `isScrollable`, `CONTAINER_ID`, or `<BackToTop />` to view files — PageLayout handles scrolling and back-to-top automatically
- **Do not** use `className="!p-0 !gap-0"` on PageSection — use the `flush` prop instead
- **Do not** hand-build a page-skeleton block or full-page spinner — render chrome real and drive regions off `loading` primitives
- **Do not** set the title on a layout/`Outlet` wrapper, and **do not** ship a routed view without a title — the leaf view owns `document.title`; a missing one leaves the title stale from the previous page
- **Do not** put the org name in a title segment, and **do not** interpolate `${x?.name}` into a `title` string (prints `"undefined"`) — use `segments`, which drops unset values

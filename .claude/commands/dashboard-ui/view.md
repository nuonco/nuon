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
   export const MyChildPage = () => (
     <PageSection>
       {/* content */}
     </PageSection>
   )
   ```

   **Detail page with flush header** (inside App/Install layout):
   ```tsx
   export const MyDetailPage = () => (
     <>
       <PageSection flush>
         <MyHeader />
       </PageSection>
       <PageSection>
         {/* content */}
       </PageSection>
     </>
   )
   ```

   Scrolling, BackToTop, and SubNav sticky are all handled automatically by PageLayout — do not add them manually.

## Anti-Patterns

- **Do not** register an install-level route outside `InstallLayout.children` — the view will render without its providers
- **Do not** omit the `enabled` guard on `useQuery` — `org` and `install` are `undefined` on the first render before providers hydrate, causing "Cannot read properties of undefined"
- **Do not** add `SurfacesProvider` or `ToastProvider` in a child view — `InstallLayout` already provides them
- **Do not** call `useInstall()` outside a route that is a child of `InstallLayout` — the provider won't be present
- **Do not** add `isScrollable`, `CONTAINER_ID`, or `<BackToTop />` to view files — PageLayout handles scrolling and back-to-top automatically
- **Do not** use `className="!p-0 !gap-0"` on PageSection — use the `flush` prop instead

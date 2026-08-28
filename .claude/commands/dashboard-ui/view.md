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

8. Use the scaffolds for the page shell and heading row — never hand-assemble `PageLayout` / `PageHeader` / `HeadingGroup` (UXDR 020, plan 024). `ListPage` for a page listing one resource, `DetailPage` for a detail/run/document page, `SectionHeader` or `DetailHeader` for the heading row. `variant="page"` at the top of a route tree (owns the `PageLayout`); the default `variant="section"` for anything mounted via a parent layout's `Outlet`.

   **`SectionHeader` or `DetailHeader`?** Does the header identify a resource — a resource ID, `BackLink`, label badges, a status chip, timestamps, or a metadata block? → identity header → **`DetailHeader`**. A heading that just names what you're looking at ("Components", "Install state", "Processes") → **`SectionHeader`**. `DetailHeader` renders `SectionHeader`'s row internally, so the heading row itself is identical either way.

   **Org-level list page** (top of a route tree):
   ```tsx
   export const MyPage = () => (
     <>
       <PageTitle title="My page" />
       <Breadcrumbs breadcrumbs={[...]} />
       <ListPage
         variant="page"
         title="My page"
         description="What this page is for."
         createAction={<CreateThingButton variant="primary" />}
       >
         <ThingsTable shouldPoll />
       </ListPage>
     </>
   )
   ```

   **Child list page inside App/Install/Settings layout** (no PageLayout):
   ```tsx
   export const MyChildPage = () => {
     const { install } = useInstall()
     return (
       <>
         <PageTitle segments={['My page', install?.name]} />
         <Breadcrumbs breadcrumbs={[...]} />
         <ListPage title="My page" description="What this page is for.">
           <ThingsTable shouldPoll />
         </ListPage>
       </>
     )
   }
   ```

   **Non-list section page** (document, config page, tab layout) — `SectionHeader` inside a `PageSection`:
   ```tsx
   export const MyConfigPage = () => (
     <PageSection>
       <PageTitle segments={['Configuration', app?.name]} />
       <Breadcrumbs breadcrumbs={[...]} />
       <SectionHeader title="Configuration" description="What this page shows." actions={<EditButton />} />
       {/* content */}
     </PageSection>
   )
   ```

   **Detail page** (inside App/Install layout) — `DetailPage` + `DetailHeader`. Metadata always goes in the `metadata` slot, which renders a `Card` of `LabeledValue`/`LabeledStatus` below the heading row (no inline top-right block, no count threshold):
   ```tsx
   export const MyDetailPage = () => {
     const { install } = useInstall()
     return (
       <>
         <PageTitle segments={[resource?.name, install?.name]} />
         <Breadcrumbs breadcrumbs={[...]} />
         <DetailPage
           header={
             <DetailHeader
               title={resource?.name}
               id={resource?.id}
               loading={isLoading}
               actions={<ManagementDropdown />}
               metadata={
                 <>
                   <LabeledStatus label="Status" statusProps={{ status: resource?.status_v2?.status }} />
                   <LabeledValue label="Created"><Time time={resource?.created_at} format="relative" /></LabeledValue>
                 </>
               }
             />
           }
           banners={resource?.composite_error ? <CompositeError error={resource.composite_error} /> : null}
         >
           {/* content */}
         </DetailPage>
       </>
     )
   }
   ```

   **Run page** (a thing that executed) — same scaffold plus routed `TabNav`. Landing tab is always **Summary** (`RunSummary` from `@/components/runs/RunSummary`), then Logs · Trace · component-type tabs. Never unrouted `Tabs` for page structure:
   ```tsx
   <DetailPage header={<MyRunHeader />} tabNav={{ basePath, tabs }}>
     <Outlet />
   </DetailPage>
   ```

   **Entity page** (a configured thing) — sections in the main column, related history in a `HistoryRail`, with a `HistoryPanelButton` in the header's `actions` for narrow widths. It graduates to routed `TabNav` when the page gains a third independent concern:
   ```tsx
   const history = <RunTimeline … shouldPoll />

   <DetailPage
     header={<DetailHeader backLink={false} title="Sandbox details" id={sandbox?.id}
       actions={<><HistoryPanelButton title="Sandbox history" history={history} /><ManagementDropdown /></>} />}
   >
     <HistoryRail title="Sandbox history" history={history}>
       <SandboxConfigCard config={config} />
     </HistoryRail>
   </DetailPage>
   ```

   `PageTitle` and `Breadcrumbs` are headless setters — render them as siblings before the scaffold, not inside it.

   Search, pagination and filters belong to `Table`/`Timeline` (`enableSearch`, `pagination`, `filterActions`), not to the page — `ListPage` has no slot for them.

   Scrolling, BackToTop, and SubNav sticky are all handled automatically by PageLayout — do not add them manually.

## List bodies & drill-down (keystone 010, plan 025)

A list page's body is `Table`, `Timeline`, or `Cards`, and the page gets exactly **one**
disclosure mechanism.

- **`Table`** (default) — the items *exist* and you compare them by attribute. If you can name
  three column headers, it is a table.
- **`Timeline`** — the items *happened*: primary sort is `created_at` and every row has a status.
- **`Cards`** — only when the item's content is the point and nothing compares column-wise. Needs
  a written exception; "it looks nicer as cards" is not one.

Drill-down: own lifecycle (actions, logs, shareable link) → its own **page**; read-only inspection
where list context matters → **panel**; a few scannable lines with no sub-structure → **expand**.
If the disclosed content has a diff, a nested collection, tabs, or its own actions, it is not an
expand.

Panels copy `InstallResourceDetailPanel`'s shape (`size="half"`, plain heading, status + `ID` row,
2-col `LabeledValue` grid, `Divider dividerWord="…"` sections) and open from the row's identity
cell using `panelTriggerClass` from `components/surfaces/panel-trigger.ts` — never a whole-row
click or a trailing "Details" button.

Sanctioned exceptions (leave them alone): announcements cards, `BranchCards`, and the
`ActiveWorkflows` cards above the install Workflows timeline (kept by explicit team request).

See `DESIGN.md` §5 "List bodies & drill-down".

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
- **Do not** hand-assemble a heading row (`PageHeader` + `HeadingGroup` + an actions `div`, or a bare heading `Text`) — use `SectionHeader`/`ListPage`
- **Do not** hand-roll a detail page's identity header (`BackLink` + heading `Text` + `<ID>` + a metadata row) — use `DetailHeader`
- **Do not** hand-roll a history rail (`@container` + `grid-cols-12` + a `@5xl:hidden` panel button) — use `HistoryRail` + `HistoryPanelButton`
- **Do not** use unrouted `Tabs` for a detail page's structure — detail-page tabs are always routed `TabNav`, and a run page's landing tab is Summary
- **Do not** render a run page without a Summary tab, or put a run's metadata anywhere but `DetailHeader`'s `metadata` slot
- **Do not** render `PageLayout` from a view mounted via a parent layout's `Outlet` — use the default `variant="section"`
- **Do not** put metadata (IDs, timestamps, badges, grids) in a second header column — it is content below the heading row
- **Do not** leave a create button in the table's `filterActions` or only in the empty state — a UI-creatable list gets one create button in `ListPage`'s `createAction`
- **Do not** add a search or pagination control to the page — `Table`/`Timeline` own those
- **Do not** render a resource collection as cards, or mix cards/expands/panels on one page — pick one body by the test above and one disclosure mechanism
- **Do not** put a diff, a nested collection, or row actions inside an `Expand` — that content belongs in a panel or its own page
- **Do not** add a view switcher (table ⇄ cards) — one rendering per list
- **Do not** hand-build a page-skeleton block or full-page spinner — render chrome real and drive regions off `loading` primitives
- **Do not** set the title on a layout/`Outlet` wrapper, and **do not** ship a routed view without a title — the leaf view owns `document.title`; a missing one leaves the title stale from the previous page
- **Do not** put the org name in a title segment, and **do not** interpolate `${x?.name}` into a `title` string (prints `"undefined"`) — use `segments`, which drops unset values

# Next.js to React SPA Migration Status

## Overview
This document tracks the progress of migrating the Nuon dashboard from Next.js (app directory) to a React SPA using React Router.

## ✅ Completed Work

### Critical Infrastructure
1. **SPA Entry Point** - `src/spa-entry.tsx`
   - Authentication bootstrap
   - Account and user data fetching
   - Provider hierarchy setup

2. **Core Layouts** - All migrated to SPA
   - ✅ `OrgLayout` - Organization-level layout with 8 providers
   - ✅ `AppLayout` - App-level layout with context
   - ✅ `InstallLayout` - Install-level layout with context
   - ✅ `InstallActionRunLayout` - NEW - Action run layout with tabs

3. **Router Configuration** - `src/routes/index.tsx`
   - React Router with nested routing
   - Lazy-loaded pages with Suspense
   - All core routes configured including action run routes

### Fixed Critical Issues
1. **✅ AppsPage** - Was showing empty, now fetches data with `usePolling`
2. **✅ InstallsPage** - Was showing empty, now fetches data with `usePolling`
3. **✅ InstallOverview** - Fully migrated with README and Current Inputs sections

### New Pages Created
1. ✅ `InstallActionRunSummary` - Action run summary with step graph and outputs
2. ✅ `InstallActionRunLogs` - Action run logs with log streaming

## 📊 Migration Progress by Section

### Org-Level Pages (5 pages)
- ✅ OrgDashboard - Redirects to /apps
- ✅ AppsPage - Apps list with data fetching
- ✅ InstallsPage - Installs list with data fetching
- ⏳ OrgRunner - Placeholder
- ⏳ TeamPage - Placeholder

### App-Level Pages (11 pages)
- ⏳ AppOverview - Placeholder (needs migration)
- ⏳ AppComponents - Placeholder
- ⏳ AppComponentDetail - Placeholder
- ⏳ AppInstalls - Placeholder
- ⏳ AppActions - Placeholder
- ⏳ AppActionDetail - Placeholder
- ⏳ AppPolicies - Placeholder
- ⏳ AppPolicyDetail - Placeholder
- ⏳ AppReadme - Placeholder
- ⏳ AppRoles - Placeholder

### Install-Level Pages (15 pages)
- ✅ InstallOverview - Fully migrated
- ⏳ InstallComponents - Placeholder
- ⏳ InstallComponentDetail - Placeholder
- ⏳ InstallWorkflows - Placeholder
- ⏳ InstallWorkflowDetail - Placeholder
- ⏳ InstallActions - Placeholder
- ⏳ InstallActionDetail - Placeholder
- ✅ InstallActionRunSummary - NEW, fully implemented
- ✅ InstallActionRunLogs - NEW, fully implemented
- ⏳ InstallRunner - Placeholder
- ⏳ InstallSandbox - Placeholder
- ⏳ InstallSandboxRun - Placeholder
- ⏳ InstallPolicies - Placeholder
- ⏳ InstallRoles - Placeholder
- ⏳ InstallStacks - Placeholder

### Root-Level Pages
- ✅ HomePage - Basic structure
- ❌ OnboardingPage - Not created
- ❌ RequestAccessPage - Not created

**Total Progress: 8/36 pages fully migrated (22%)**

## 🎯 Migration Pattern

### Next.js Pattern (Server-Side)
```typescript
// layout.tsx
export default async function Layout({ children, params }) {
  const { 'org-id': orgId } = await params
  const { data } = await getServerSideData({ orgId })
  
  return (
    <Provider initData={data}>
      {children}
    </Provider>
  )
}

// page.tsx with server component wrapper
export default async function Page({ params, searchParams }) {
  const sp = await searchParams
  return (
    <Suspense fallback={<Skeleton />}>
      <DataComponent orgId={orgId} offset={sp['offset']} />
    </Suspense>
  )
}

// server component (data-component.tsx)
export async function DataComponent({ orgId, offset }) {
  const { data } = await fetchData({ orgId, offset })
  return <UI data={data} />
}
```

### SPA Pattern (Client-Side)
```typescript
// Layout.tsx
export default function Layout() {
  const { orgId } = useParams()
  
  const { data, isLoading } = usePolling({
    path: `/api/orgs/${orgId}/resource`,
    shouldPoll: true,
    pollInterval: 30000,
  })
  
  if (isLoading) return <LoadingSpinner />
  
  return (
    <Provider initData={data}>
      <Outlet />
    </Provider>
  )
}

// Page.tsx - inline data fetching
export default function Page() {
  const { orgId } = useParams()
  const [searchParams] = useSearchParams()
  const offset = searchParams.get('offset') || '0'
  
  const { data, isLoading, error } = usePolling({
    path: `/api/orgs/${orgId}/resource?offset=${offset}`,
    shouldPoll: true,
    pollInterval: 30000,
  })
  
  if (isLoading) return <LoadingState />
  if (error) return <ErrorState error={error} />
  
  return <UI data={data} />
}
```

### Key Differences
1. **Params**: `await params` → `useParams()`
2. **Search Params**: `await searchParams` → `useSearchParams()`
3. **Data Fetching**: Server `await getData()` → Client `usePolling()`
4. **Child Rendering**: `{children}` → `<Outlet />`
5. **Loading States**: `<Suspense>` → Explicit `isLoading` checks
6. **Error Handling**: `<ErrorBoundary>` → Explicit `error` checks

## 🔧 Common Implementation Details

### API Response Structure
```typescript
// usePolling returns:
{
  data: T | null,           // The actual data
  error: TAPIError | null,  // Error object if failed
  isLoading: boolean,       // Loading state
  headers: Record<string, string> | null, // Response headers
  status: number | null     // HTTP status
}
```

### Pagination Pattern
```typescript
const { data: response, headers } = usePolling<TItem[]>({
  path: `/api/orgs/${orgId}/items?limit=10&offset=${offset}`,
  shouldPoll: true,
})

const pagination = {
  limit: Number(headers?.['x-nuon-page-limit'] ?? 10),
  hasNext: headers?.['x-nuon-page-next'] === 'true',
  offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
}
```

### Server Component to Client Component
When migrating server components that fetch data:

**Before (Server Component):**
```typescript
// apps-table.tsx
export async function AppsTable({ orgId, offset }) {
  const { data: apps } = await getApps({ orgId, offset })
  return <Table apps={apps} />
}
```

**After (Inline in Page):**
```typescript
// AppsPage.tsx
export default function AppsPage() {
  const { orgId } = useParams()
  const { data: apps } = usePolling<TApp[]>({
    path: `/api/orgs/${orgId}/apps?offset=${offset}`,
    shouldPoll: true,
  })
  return <Table apps={apps || []} />
}
```

## 🚀 Next Steps

### Immediate Priorities
1. Migrate `AppOverview` page (complex, has multiple sections)
2. Migrate remaining high-traffic pages:
   - `AppComponents` + `AppComponentDetail`
   - `InstallComponents` + `InstallComponentDetail`
   - `InstallActions` + `InstallActionDetail`

### Medium Priority
3. Migrate remaining App pages (7 pages)
4. Migrate remaining Install pages (9 pages)
5. Create `OnboardingPage` and `RequestAccessPage`

### Final Steps
6. Test all migrated pages thoroughly
7. Remove Next.js app directory (`src/app/`)
8. Remove Next.js dependencies from `package.json`
9. Update build configuration
10. Update documentation

## 📝 Testing Checklist

For each migrated page:
- [ ] Run `touch ~/.nuonctl-restart-dashboard-ui`
- [ ] Navigate to page in Chrome
- [ ] Verify data loads correctly
- [ ] Verify pagination works (if applicable)
- [ ] Verify search/filter works (if applicable)
- [ ] Verify navigation (breadcrumbs, tabs, links)
- [ ] Verify error states display properly
- [ ] Check console for errors

## 🐛 Known Issues

None currently.

## 📚 References

- Plan: `/Users/jonmorehouse/.claude/plans/lovely-tickling-starfish.md`
- Next.js App: `src/app/` (reference for migration)
- SPA Pages: `src/pages/`
- Layouts: `src/pages/layouts/`
- Routes: `src/routes/index.tsx`

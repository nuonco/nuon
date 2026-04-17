# Dashboard UI Service

The **Dashboard UI** is the primary web application frontend for the Nuon platform. It is a **Go BFF (Backend-for-Frontend) + React SPA** architecture.

## Architecture Overview

```
services/dashboard-ui/
├── client/     ← React SPA (ALL new work goes here)
└── server/     ← Go BFF (Gin + Uber fx)
```

- **`client/`** — React SPA built with React Router v7, TanStack Query, Tailwind CSS, ESBuild
- **`server/`** — Go BFF that serves the SPA, validates auth cookies, injects runtime config, and provides streaming API handlers

## Go BFF Server (`server/`)

The Go server (Gin + Uber fx) handles:
- Serving the compiled SPA from `dist/`
- Auth middleware: validates the cookie set by the external auth service
- Runtime config injection: writes `window.__NUON_CONFIG__` into the HTML before serving
- **Reverse proxy**: all `/v1/*` requests from the SPA are forwarded to ctl-api — the BFF extracts the `X-Nuon-Auth` cookie server-side and sets `Authorization: Bearer <token>` so the browser never needs to send the cookie cross-domain
- Streaming API handlers (e.g., log streaming, log download)

### BFF API Endpoints (`server/internal/handlers/`)

The BFF exposes its own `/api/*` endpoints (separate from the `/v1/*` reverse proxy). These handlers authenticate via the `X-Nuon-Auth` cookie and create a nuon-go client server-side.

**Log streams** (`log_streams.go`):
- `GET /api/orgs/:orgId/log-streams/:logStreamId/logs/sse` — SSE streaming endpoint for real-time logs
- `GET /api/orgs/:orgId/log-streams/:logStreamId/logs/download` — Download logs as a text file
  - `?job_output=true` — Filter to job output only (keeps only logs with `ScopeName == "oteljob"`)

**User vs internal logs**: The runner emits logs with two OTEL scope names — `oteljob` for job execution output (builds, deploys, actions) and `system` for internal runner logs. The `user_output=true` filter keeps only records where `ScopeName == "oteljob"`.

## Client SPA (`client/`)

### Directory Structure

```
client/
├── components/         ← Reusable UI components (organized by domain)
│   ├── common/         ← Core primitives: Button, Card, Badge, Text, Modal, Toast
│   ├── layout/         ← Page structure: PageLayout, PageContent, PageSection
│   ├── navigation/     ← SubNav, Breadcrumbs, MainNav
│   ├── surfaces/       ← Modal/Panel system
│   └── [domain]/       ← Feature components (actions, workflows, runners, installs, etc.)
├── hooks/              ← Custom React hooks (47+ hooks for state and utilities)
├── lib/
│   ├── api.ts          ← Fetch wrapper (returns T directly, throws TAPIError on failure)
│   └── ctl-api/        ← Domain-specific API functions (organized by resource)
│       ├── accounts/
│       ├── apps/
│       ├── installs/
│       ├── runners/
│       ├── workflows/
│       └── ...
├── providers/          ← React context providers
├── types/
│   ├── ctl-api.types.ts       ← Extracted API types (T prefix)
│   ├── dashboard.types.ts     ← Custom types (TAPIError, etc.)
│   └── nuon-oapi-v3.d.ts      ← Auto-generated OpenAPI types (do not import directly)
├── views/              ← Page-level view components (mirrors route tree)
└── main.tsx            ← App entry point
```

## Views vs Components (Strict Separation)

- **`views/`** contains **only**: page-level view components (route content), layout components (providers/breadcrumbs/tab nav), and route orchestration.
- **`views/`** must **never** contain: modals, tables, reusable sub-components, action buttons, or any component meant to be consumed by a view.
- All feature components belong in `client/components/[domain]/`. If a `components/[domain]/` directory doesn't exist yet, create it.

## Layout System

The layout system handles page structure, scrolling, and back-to-top automatically. Pages assemble from clear building blocks without worrying about scroll containers, overflow, or positioning.

### Layout Component Hierarchy

```
MainLayout (flex row: sidebar + content)
├── MainSidebar (desktop: static flex child, mobile: fixed overlay)
├── Mobile backdrop
└── Content wrapper (flex-1, flex-col, overflow-hidden)
    ├── PageLayout (flex-1, contains topbar + scroll container)
    │   ├── MainTopbar
    │   └── Scroll container (overflow-y-auto, auto BackToTop)
    │       ├── PageHeader (optional)
    │       ├── PageContent (flex direction: column or row)
    │       │   ├── SubNav (optional, sticky on desktop)
    │       │   └── Page content / Outlet
    │       └── BackToTop (automatic, sticky bottom-right)
    └── OrgStatusBar (flex-none, pinned at bottom)
```

### Building Pages

**Org-level page** (top-level route like Apps, Installs, Team):
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

**Child page inside App/Install layout** (rendered via `<Outlet />`):
```tsx
export const MyChildPage = () => (
  <PageSection>
    {/* content — that's it */}
  </PageSection>
)
```

**Detail page with flush header**:
```tsx
export const DeployDetail = () => (
  <>
    <PageSection flush>
      <DeployHeader />
    </PageSection>
    <PageSection>
      <Logs />
    </PageSection>
  </>
)
```

### Layout Components

| Component | Purpose | Key Props |
|-----------|---------|-----------|
| `PageLayout` | Top-level page wrapper. Renders topbar, scroll container, and BackToTop automatically. | `variant` (`dashboard-page` / `single-page`), `hideBreadcrumbs` |
| `PageContent` | Sets flex direction for content area. | `variant` (`column` default, `row` for SubNav layouts) |
| `PageSection` | Content block with standard padding/gap. | `flush` (removes padding/gap for full-bleed content) |
| `PageHeader` | Page heading area above content. | Standard div props |
| `SubNav` | Secondary navigation sidebar. Sticky on desktop, horizontal scroll on mobile. | `basePath`, `links` |

### What You Get For Free

- **Scrolling**: PageLayout's inner div is always the scroll container (`overflow-y-auto`)
- **Back to top**: Auto-rendered inside PageLayout, appears after 400px scroll
- **SubNav sticky**: Stays pinned on desktop while content scrolls beside it
- **OrgStatusBar**: Pinned at the bottom, outside the scroll area

### Do NOT

- Add `isScrollable` to any component — it's ignored (kept for backwards compat only)
- Create `CONTAINER_ID` constants or pass `id` props to scroll containers
- Import or render `<BackToTop />` in view files — PageLayout handles it
- Use `className="!p-0 !gap-0"` on PageSection — use the `flush` prop instead

### Mobile Sidebar

The main sidebar uses a fixed overlay on mobile (`w-[280px]`, slides in from left) with a backdrop. On desktop it's a normal flex child with a collapsible width transition (Alt+S).

## Routing

Routing uses **React Router v7** with nested routes. Route files live in `client/views/` mirroring the URL hierarchy.

Example route structure:
```
/:orgId/installs/:installId/
├── actions/:actionId/
│   └── runs/:actionRunId/    ← ActionRunLayout wraps with provider + breadcrumbs + TabNav
│       ├── (summary tab)
│       └── (logs tab)
```

## API Integration (`client/lib/api.ts`)

### Return Type Behavior

`api<T>()` returns `T` directly — **not `{ data: T }`**. It throws `TAPIError` on failure.

```typescript
// ✅ Correct
const runner = await getRunner({ runnerId, orgId })
runner.id  // direct access

// ❌ Wrong — there is no .data wrapper in the client SPA
const { data: runner } = await getRunner({ runnerId, orgId })
```

### API Function Pattern

Functions in `client/lib/ctl-api/` follow this pattern:

```typescript
// GET function
export const getRunner = ({
  runnerId,
  orgId,
}: {
  runnerId: string
  orgId: string
}) =>
  api<TRunner>({
    path: `runners/${runnerId}`,
    orgId,
  })

// POST function with body
export async function createInstallConfig({
  body,
  installId,
  orgId,
}: {
  body: TCreateInstallConfigBody
  installId: string
  orgId: string
}) {
  return api<TInstallConfig>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/configs`,
  })
}
```

Each domain directory has an `index.ts` barrel export. Import from `@/lib`:

```typescript
import { getRunner, createInstallConfig } from '@/lib'
```

### Error Handling

`api()` throws `TAPIError` on non-2xx responses. Catch in `useMutation` `onError` or use TanStack Query's error state.

## State Management

### Provider Hierarchy

```
ConfigProvider
└── QueryClientProvider
    └── AuthProvider
        └── APIHealthProvider
            └── (layout providers per page)
                ├── InstallProvider
                ├── ToastProvider
                └── SurfacesProvider
```

### TanStack Query Patterns

**Data fetching (`useQuery`)**:
```typescript
const { data: runner, isLoading, error } = useQuery({
  queryKey: ['runner', runnerId],
  queryFn: () => getRunner({ runnerId, orgId }),
  enabled: !!runnerId,
  refetchInterval: shouldPoll ? 5000 : false,
})
```

**Mutations (`useMutation`)**:
```typescript
const { mutate: cancel, isPending } = useMutation({
  mutationFn: ({ workflowId }: { workflowId: string }) =>
    cancelWorkflow({ workflowId, orgId }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['workflows'] })
    removeModal(modalId)
  },
  onError: (err: TAPIError) => {
    showErrorToast(err.error)
  },
})
```

### Custom Hooks

Access providers through custom hooks — never use `useContext` directly:

```typescript
const { org } = useOrg()
const { account } = useAccount()
const { install } = useInstall()
const config = useConfig()
const { addModal, removeModal } = useSurfaces()
```

## Authentication

The external auth service sets a `X-Nuon-Auth` httponly cookie scoped to the app domain. The Go BFF validates the cookie on page loads and extracts the token server-side when reverse-proxying `/v1/*` API requests — the browser never sends the cookie to ctl-api directly. On the client, `AuthProvider` calls `getMe()` at startup to load the current account. On 401 API responses, `api.ts` automatically redirects to the login page.

## Runtime Config (`useConfig()`)

The Go server injects environment variables into the HTML as `window.__NUON_CONFIG__` before serving the SPA. Access config values via `useConfig()`:

```typescript
const config = useConfig()
const apiUrl = config.apiUrl
const datadogEnabled = config.datadogEnabled
```

Never hardcode API URLs or feature flags — always read from config.

## TypeScript Conventions

### Naming

- **`T` prefix** — data types and API response types: `TApp`, `TInstall`, `TRunner`
- **`I` prefix** — component props and configuration interfaces: `IModal`, `IButton`

### Type Imports

**Always import from `ctl-api.types.ts`** — never directly from the generated `nuon-oapi-v3.d.ts`:

```typescript
// ✅ Correct
import type { TApp, TInstall, TRunner } from '@/types/ctl-api.types'

// ❌ Wrong — never import directly from generated file
import type { components } from '@/types/nuon-oapi-v3'
type TApp = components['schemas']['app.App']
```

To add new types, extract them in `ctl-api.types.ts`:
```typescript
// In /client/types/ctl-api.types.ts
export type TNewResource = components['schemas']['app.NewResource']
```

## Component Patterns

### Always Check Existing Components First

Before building a new component, **check `client/components/common/` and other domain directories** for an existing component that meets your needs.

**Stories files (`.stories.tsx`) are the primary reference for how to use a component.** They contain live examples of correct prop usage, edge cases, and patterns. Always read a component's stories file before using or modifying it — this is faster and more reliable than inferring usage from the TypeScript interface alone.

```bash
# Find stories for a component
glob pattern: client/components/**/*.stories.tsx
```

### Icons

**Use the `Icon` component (`client/components/common/Icon.tsx`) for ALL icons.** Never import from `lucide-react`, `heroicons`, or any other icon package. The project uses `@phosphor-icons/react` via the `Icon` wrapper — all icon usage must go through it.

```tsx
// ✅ Correct
import { Icon } from '@/components/common/Icon'
<Icon variant="MagnifyingGlassIcon" size={16} />

// ❌ Wrong — do not import icons from other packages
import { Search } from 'lucide-react'
import { MagnifyingGlassIcon } from '@heroicons/react/24/outline'
```

Browse Phosphor icons at https://phosphoricons.com. The variant name is the Phosphor export name (e.g., `MagnifyingGlass` → `MagnifyingGlassIcon`). Custom icons for cloud providers and tools are also available — see the `customIcons` map in `Icon.tsx`.

### `Tabs` Component — Key Casing

The `Tabs` component renders tab labels by running each object key through `toSentenceCase(camelToWords(key))`. `toSentenceCase` capitalizes the first character and **lowercases everything else**. Always write tab keys in all-lowercase so the rendered label is correct:

```tsx
// ✅ Correct — keys are all-lowercase, rendered as "Create your own app" / "Demo using a sample app"
<Tabs tabs={{ 'create your own app': <CustomTab />, 'demo using a sample app': <DemoTab /> }} />

// ❌ Wrong — title case keys render incorrectly: "Create your own app" loses capitals mid-string
<Tabs tabs={{ 'Create Your Own App': <CustomTab /> }} />
```

### Container / Component Pattern

Feature components use a **container/component split** to separate data-fetching from presentation. Every feature component directory follows this structure:

```
client/components/[domain]/MyComponent/
├── MyComponent.tsx              ← Pure presentational component (props in, JSX out)
├── MyComponentContainer.tsx     ← Data-fetching wrapper (hooks, queries, mutations)
├── MyComponent.stories.tsx      ← Ladle stories (required)
├── index.ts                     ← Barrel export
```

**`MyComponent.tsx`** — The presentational component. Receives all data as props. No `useQuery`, `useMutation`, or context hooks that require providers. This is the component that stories render directly.

**`MyComponentContainer.tsx`** — The container. Calls hooks (`useOrg()`, `useQuery()`, etc.) and passes resolved data to the presentational component. Views and other containers import this via the barrel.

**`index.ts`** — Barrel export. Exports the container as the default/primary export, and the presentational component as a named export:
```typescript
export { MyComponentContainer as MyComponent } from './MyComponentContainer'
export { MyComponent as MyComponentComponent } from './MyComponent'
```

**When to use this pattern**: Any component that calls context hooks (`useOrg`, `useInstall`, `useDeploy`, etc.) or TanStack Query hooks. Simple presentational components (Button, Badge, etc.) stay as flat files.

**Important**: Never have both a flat file `MyComponent.tsx` and a directory `MyComponent/` at the same level — the flat file shadows the directory's `index.ts` and causes import resolution bugs.

### File Organization

**Flat files (for simple presentational components)**:
```
client/components/common/
├── Button.tsx
├── Badge.tsx
└── Text.tsx
```

**Directory structure (for feature components with container/component split)**:
```
client/components/[domain]/MyComponent/
├── MyComponent.tsx
├── MyComponentContainer.tsx
├── MyComponent.stories.tsx
└── index.ts
```

### Ladle Stories (Required)

Every component directory must include a `.stories.tsx` file. Stories are written for **Ladle v5** — not Storybook.

**Story format** — plain function exports only. Ladle does NOT support `StoryObj` with `render:`:
```tsx
// ✅ Correct — Ladle v5 format
export default {
  title: 'Domain/MyComponent',
}

import { MyComponent } from './MyComponent'

export const Default = () => <MyComponent items={mockItems} />
export const Empty = () => <MyComponent items={[]} />
```

```tsx
// ❌ Wrong — Storybook syntax, breaks Ladle ("got: object" error)
import type { Meta, StoryObj } from '@ladle/react'
export const Default: StoryObj = { render: () => <MyComponent /> }
```

**Stories render the presentational component**, not the container. Pass all data as props — no provider dependencies needed.

**When a component needs a context provider** (because it renders a child that calls a hook), mock the context in the story:
```tsx
import { SomeContext } from '@/providers/some-provider'

const mockValue = { /* mock context shape */ }

export const Default = () => (
  <SomeContext.Provider value={mockValue}>
    <MyComponent />
  </SomeContext.Provider>
)
```

**Modal stories** — use the `ModalStory` helper from `@/components/__stories__/helpers`:
```tsx
import { ModalStory } from '@/components/__stories__/helpers'
import { MyModal } from './MyModal'

export const Default = () => (
  <ModalStory>
    <MyModal someData={mockData} />
  </ModalStory>
)
```

**Timeline stories** — mock items must have unique `created_at` timestamps on different calendar days. The `Timeline` component groups by date, so duplicate dates cause React key warnings.

**Ladle provides a `MemoryRouter`** globally — never wrap stories in another `MemoryRouter` or you'll get "cannot render a `<Router>` inside another `<Router>`".

### Modal and Panel Components

Always use `Modal` and `Panel` from `client/components/surfaces/` — never use `ModalBase` or `PanelBase` directly.

The standard pattern is **two components**: a Modal/Panel component and a Button component:

```typescript
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { useSurfaces } from '@/hooks/use-surfaces'

interface IDeleteModal extends IModal {
  item: TItem
}

export const DeleteModal = ({ item, ...props }: IDeleteModal) => {
  const { removeModal } = useSurfaces()
  const { mutate: doDelete, isPending } = useMutation({
    mutationFn: () => deleteItem({ itemId: item.id }),
    onSuccess: () => removeModal(props.modalId),
  })

  return (
    <Modal
      heading="Delete Item"
      primaryActionTrigger={{
        children: isPending ? 'Deleting...' : 'Delete',
        disabled: isPending,
        onClick: () => doDelete(),
        variant: 'danger',
      }}
      {...props}
    >
      <Text>Are you sure you want to delete {item.name}?</Text>
    </Modal>
  )
}

export const DeleteButton = ({ item, ...props }: { item: TItem } & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteModal item={item} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Delete
    </Button>
  )
}
```

**Rules**:
- Always `{...props}` spread onto `Modal`/`Panel` — never destructure `modalId`, `isVisible`, `onClose` manually
- Create the modal instance before passing to `addModal`: `const modal = <MyModal />` then `addModal(modal)`
- Close modals on success via `removeModal(props.modalId)`

## Text & Copy Style

**Always use sentence case, never title case.** This applies to all UI text: headings, buttons, labels, tab labels, empty states, tooltips, and any other copy.

- ✅ "Create your org" / "Connect a cloud account" / "Generate random name"
- ❌ "Create Your Org" / "Connect A Cloud Account" / "Generate Random Name"

The only exceptions are proper nouns (AWS, Nuon, Terraform, etc.) and acronyms.

## Dates, Times & Durations

**Always use [Luxon](https://moment.github.io/luxon/) for date/time operations.** Never use raw `Date` objects or manual millisecond math.

**Use the existing components for rendering:**

- **`<Time>`** (`client/components/common/Time.tsx`) — Renders timestamps. Supports `format="relative"` (e.g., "2 hours ago" with tooltip), `"short-datetime"`, `"long-datetime"`, `"time-only"`, `"log-datetime"`.
- **`<Duration>`** (`client/components/common/Duration.tsx`) — Renders durations between two times. Pass `beginTime` and optionally `endTime` (defaults to now). Supports `durationUnits`, `unitDisplay`, and `format` props.

```tsx
// ✅ Correct — use Time and Duration components
<Time variant="subtext" time={item.created_at} format="relative" />
<Duration variant="subtext" beginTime={process.started_at} durationUnits={['hours', 'minutes']} />

// ❌ Wrong — manual date formatting
const diffMs = Date.now() - new Date(dateStr).getTime()
const minutes = Math.floor(diffMs / (1000 * 60))
return `${minutes} minutes ago`
```

**For utility functions** that need date logic (not rendering), use Luxon's `DateTime` and `Duration` classes directly. Place shared helpers in `client/utils/time-utils.ts`.

Do not add comments unless the logic is genuinely non-obvious. Never write comments that just describe what the code does (no "// loop through items", "// close modal", "// fetch data" style comments). Let clear naming and structure document the code.

## Key Scripts

```bash
npm run dev            # Development: esbuild watch + PostCSS + BrowserSync
npm run build          # Production build (minified)
npm run build:js       # Build JS only
npm run build:css      # Build CSS only
npm run lint           # ESLint for the SPA
npm run tsc            # TypeScript type check
npm run dev:ladle      # Ladle component stories
npm test               # Vitest tests
```

**Do NOT run build commands** (`build`, `build:js`, `build:css`) unless explicitly asked. A dev process (nctl) is already running that handles builds automatically.


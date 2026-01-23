# Dashboard UI Service  

The **Dashboard UI** is the primary web application frontend for the Nuon platform, providing a comprehensive interface for managing applications, deployments, and infrastructure.

## Architecture

- **Framework**: Next.js 15+ with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS with custom design system
- **State Management**: React hooks with provider pattern
- **Authentication**: Auth0 integration via `@auth0/nextjs-auth0`
- **Testing**: Vitest with React Testing Library + MSW
- **Build Tool**: Turbo for development speed

## Authentication Patterns

### Server Actions vs API Routes for Authenticated Operations

**IMPORTANT**: For authenticated operations that need to interact with the Nuon API, use **server actions** instead of API routes.

**Server Actions (Recommended)**:
- Run in the server component context where Auth0 session access works properly
- Can directly use `auth0.getSession()` for authenticated API calls
- Follow the pattern established in `/src/actions/` directory structure
- Use `executeServerAction()` wrapper for consistent behavior and path revalidation

**API Routes (Avoid for Auth)**:
- Run in a different context where Auth0 session may not be properly accessible
- Can cause 401 authentication errors even with valid user sessions
- Should primarily be used for public endpoints or proxy functionality

**Best Practice**: 
```typescript
// ✅ Good - Modern server action pattern (src/actions/)
'use server'
import { executeServerAction } from '@/actions/execute-server-action'
import { createSomething as create } from '@/lib'

export async function createSomething({
  body,
  path,
  orgId,
}: {
  body: TCreateBody
  orgId: string
  path?: string
}) {
  return executeServerAction({
    action: create,
    args: { body, orgId },
    path,
  })
}

// ❌ Avoid - API route for authenticated operations  
export const POST = async (request: NextRequest) => {
  // Auth0 session access issues can occur here
}

// ❌ DEPRECATED - Old component-based actions (src/components/*-actions.ts)  
// These should be migrated to src/actions/ pattern
```

## API Integration & Error Handling

### Global vs Organization-Scoped Endpoints

**Critical Distinction**: Some API endpoints are "global" (account-level) and should NOT include `orgId`:

```typescript
// ✅ CORRECT - Global endpoint (no orgId)
await api<TAccount>({
  path: 'account/user-journeys/evaluation/complete',
  method: 'POST'
  // No orgId parameter
})

// ✅ CORRECT - Organization-scoped endpoint (with orgId)
await api({
  path: 'apps',
  method: 'GET',
  orgId  // Required for org-scoped endpoints
})
```

**Detection**: Check if endpoint is listed in `ctl-api/internal/middlewares/global/global.go` to determine if it needs organization context.

## API Type Generation & Mock Testing System

### Environment-Based API Testing

**Key Environment Variable**: `NUON_API_URL` determines which API the frontend consumes:

```bash
# Test against local API (developer changes)
export NUON_API_URL=http://localhost:8081
npm run generate-api-mocks
npm test

# Test against production API (default)
export NUON_API_URL=https://api.nuon.co  # or omit (this is default)
npm run generate-api-mocks
npm test
```

### Circular Reference Cleaning

**Problem**: OpenAPI specs from complex APIs contain circular references that break mock generation.

**Solution**: Surgical schema cleaning that preserves business logic while removing circular references:

```javascript
// scripts/clean-openapi-spec.js performs:
// 1. Downloads spec from NUON_API_URL/oapi/v3
// 2. Removes truly unused/internal schemas entirely
// 3. Cleans essential business schemas while preserving required fields
// 4. Outputs cleaned spec for mock generation
```

**Key Patterns**:
- **Remove unused schemas**: Internal tracking schemas not used by UI
- **Clean essential schemas**: Keep business fields (status, type, install_id) while removing circular references
- **Preserve test expectations**: Generated mocks must match what tests expect

## State Management & Provider Architecture

### Provider Pattern

The dashboard uses a hierarchical provider pattern for state management:

```typescript
// Layout hierarchy
<AccountProvider>
  <OrgProvider>
    <AppProvider>
      <InstallProvider>
        {children}
      </InstallProvider>
    </AppProvider>
  </OrgProvider>
</AccountProvider>
```

### Custom Hooks Architecture

All state access goes through custom hooks that encapsulate provider logic:

```typescript
// ✅ Good - Use custom hooks
const { account, refreshAccount } = useAccount()
const { org, switchOrg } = useOrg()
const { workflow, actions } = useWorkflow()

// ❌ Avoid - Direct context access
const account = useContext(AccountContext)  // Skip this pattern
```

### Polling Strategies

**Dynamic Polling**: Providers automatically adjust polling intervals based on state:

```typescript
// Fast polling during active operations
const useFastPolling = isWorkflowRunning || isOnboarding
const interval = useFastPolling ? 5000 : 20000

// Automatic pause/resume based on visibility
const shouldPoll = isVisible && !isComplete
```

## Component Directory Structure

### File Organization Patterns

**Flat Structure (Preferred)**: Most components should be created as flat files in their respective directories:

```
src/components/common/
├── Button.tsx
├── Button.stories.tsx
├── Badge.tsx
├── Badge.stories.tsx
├── PropertyGrid.tsx
├── PropertyGrid.stories.tsx
└── Text.tsx
```

**Directory Structure (Only When Necessary)**: Create a component directory ONLY when you have multiple sub-components and need to expose only the parent component:

```
src/components/common/EmptyState/
├── EmptyState.tsx          # Main component
├── EmptyState.stories.tsx  # Stories file
├── EmptyGraphic.tsx        # Internal sub-component
└── index.ts               # Exports only EmptyState (hides EmptyGraphic)
```

### When to Use Each Pattern

**Use Flat Structure When:**
- Component is self-contained
- No internal sub-components needed
- Component + stories is all that's required
- Most common case (90% of components)

**Use Directory Structure When:**
- Component has multiple internal sub-components
- You need to hide internal implementation details
- You want clean imports (e.g., `import { EmptyState } from './EmptyState'` instead of `import { EmptyState } from './EmptyState/EmptyState'`)
- Component has complex internal architecture

**Examples:**
- ✅ `PropertyGrid.tsx` - Flat file (self-contained grid component)
- ✅ `Button.tsx` - Flat file (simple button component)  
- ✅ `EmptyState/` - Directory (has EmptyGraphic sub-component that users shouldn't import directly)
- ❌ `PropertyGrid/PropertyGrid.tsx` - Unnecessary directory for simple component

### Import Patterns

**Flat Structure Imports:**
```typescript
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Button } from '@/components/common/Button'
```

**Directory Structure Imports:**
```typescript
import { EmptyState } from '@/components/common/EmptyState'  # Clean thanks to index.ts
```

**Key Rule**: Only create component directories when you have legitimate architectural complexity, not just for organization.

## Modal and Panel Components

### Always Use Modal and Panel (Not ModalBase or PanelBase)

**CRITICAL**: When working with modal dialogs or side panels, always use the `Modal` and `Panel` components from `/components/surfaces/`. Never use `ModalBase` or `PanelBase` directly.

**Before Implementation**: Check `Modal.stories.tsx` and `Panel.stories.tsx` for usage examples! The stories show the exact patterns you should follow.

**Why This Matters**:
- `Modal` and `Panel` are wrapper components that provide additional features
- They handle URL search param integration automatically
- They can optionally render trigger buttons
- They're the public API - `ModalBase` and `PanelBase` are internal implementation details

**Correct Usage Pattern**:

The standard pattern is to create TWO components - a Modal component and a Button component:

```typescript
// ✅ CORRECT - Modal component pattern
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { useSurfaces } from '@/hooks/use-surfaces'

// 1. Create modal component that extends IModal
interface IMyModal extends IModal {
  item: TItem
}

export const MyModal = ({ item, ...props }: IMyModal) => {
  const { removeModal } = useSurfaces()
  const { data, error, isLoading, execute } = useServerAction({
    action: deleteItem,
  })

  useServerActionToast({
    data,
    error,
    successHeading: 'Item deleted',
    onSuccess: () => {
      removeModal(props.modalId)  // Close modal on success
    },
  })

  return (
    <Modal
      heading="Delete Item"
      primaryActionTrigger={{
        children: isLoading ? 'Deleting...' : 'Delete',
        onClick: () => execute({ itemId: item.id }),
        disabled: isLoading,
        variant: 'danger',
      }}
      {...props}  // CRITICAL: Spread props to pass modalId, isVisible, onClose, etc.
    >
      <Text>Are you sure you want to delete {item.name}?</Text>
    </Modal>
  )
}

// 2. Create button component that triggers the modal
interface IMyButton extends IButtonAsButton {
  item: TItem
}

export const MyButton = ({ item, ...props }: IMyButton) => {
  const { addModal } = useSurfaces()
  const modal = <MyModal item={item} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Delete Item
    </Button>
  )
}
```

**Panel Usage** (same pattern):

```typescript
// ✅ CORRECT - Panel component pattern
import { Panel, type IPanel } from '@/components/surfaces/Panel'

interface IMyPanel extends IPanel {
  data: TData
}

export const MyPanel = ({ data, ...props }: IMyPanel) => {
  return (
    <Panel
      heading="Details"
      size="half"
      {...props}  // Spread props to pass panelId, isVisible, onClose, etc.
    >
      <Text>Panel content here</Text>
    </Panel>
  )
}

export const MyPanelButton = ({ data, ...props }: IButtonAsButton & { data: TData }) => {
  const { addPanel } = useSurfaces()
  const panel = <MyPanel data={data} />

  return (
    <Button onClick={() => addPanel(panel)} {...props}>
      View Details
    </Button>
  )
}
```

**CRITICAL MISTAKES TO AVOID**:

```typescript
// ❌ WRONG - Don't import or use ModalBase/PanelBase
import { ModalBase } from '@/components/surfaces/Modal'  // Never do this!
import { PanelBase } from '@/components/surfaces/Panel'  // Never do this!

// ❌ WRONG - Don't destructure IModal props manually
export const MyModal = ({
  item,
  isVisible,     // Don't do this
  modalId,       // Don't do this
  onClose,       // Don't do this
  modalKey,      // Don't do this
  ...props
}: IMyModal) => {
  return (
    <Modal
      isVisible={isVisible}   // Don't pass explicitly
      modalId={modalId}       // Don't pass explicitly
      {...props}
    >
  )
}

// ✅ CORRECT - Just spread all props
export const MyModal = ({ item, ...props }: IMyModal) => {
  return <Modal {...props}>  {/* Let Modal handle all its own props */}
}

// ❌ WRONG - Don't call addModal inline
<Button onClick={() => addModal(<MyModal item={item} />)}>

// ✅ CORRECT - Create modal instance first
const modal = <MyModal item={item} />
return <Button onClick={() => addModal(modal)}>
```
```

**Best Practices**:
- **Always use `Modal` and `Panel`** - never import `ModalBase` or `PanelBase`
- **Create two components**: Modal/Panel component + Button component
- **Spread props**: Always use `{...props}` to pass IModal/IPanel props to the component
- **Don't destructure surface props**: Let Modal/Panel handle `isVisible`, `modalId`, `onClose`, etc.
- **Create modal instance first**: `const modal = <MyModal />` then `addModal(modal)`
- **Close modals on success**: Use `removeModal(props.modalId)` in success callbacks

## Code Quality & Development

### ESLint Configuration Impact

**Development Challenge**: The `no-console` ESLint rule prevents debugging console statements:

```typescript
console.log('Debug info')  // ❌ ESLint error
```

**Workarounds**:
- Temporarily disable rule with `// eslint-disable-next-line no-console`
- Use browser DevTools for runtime debugging
- Implement proper error boundaries for user-facing error handling

### Component Usage Discovery

**CRITICAL FIRST STEP**: Before implementing any component, always look for `.stories.tsx` files to see real usage examples.

**Why Stories Files Are Essential**:
- Stories show the **intended usage patterns** for components
- They demonstrate **real-world examples** of prop combinations
- They reveal **hidden features** and optional props you might miss
- They show **correct patterns** used by the component authors
- They prevent you from **guessing or inventing** incorrect APIs

**Discovery Process** (follow this order):

1. **Check for .stories.tsx file FIRST**
   ```bash
   # If using Button component, look for:
   src/components/common/Button.stories.tsx

   # If using Modal component, look for:
   src/components/surfaces/Modal.stories.tsx
   ```

2. **Read the stories to see usage examples**
   - Look at how props are passed
   - See what values are used for enums/unions
   - Observe patterns like spreading props (`{...props}`)
   - Notice integration with hooks (`useSurfaces`, `useServerAction`)

3. **Only then read the component source** to verify prop types

**Example Discovery Flow**:

```typescript
// Step 1: Find the stories file
// src/components/surfaces/Modal.stories.tsx

// Step 2: Read the stories to see actual usage
const openModal = () => {
  addModal(
    <Modal
      heading="Create New Project"
      primaryActionTrigger={{
        children: 'Create Project',
        onClick: () => alert('Project created!'),
      }}
    >
      <div className="p-6">
        {/* Modal content */}
      </div>
    </Modal>
  )
}

// Step 3: Now you understand the pattern - Modal is passed to addModal()
// Step 4: Verify prop types in Modal.tsx if needed
```

**Benefits of Stories-First Approach**:
- ✅ See multiple usage examples in one place
- ✅ Understand integration patterns (hooks, state management)
- ✅ Learn component conventions (two-component patterns, prop spreading)
- ✅ Avoid guessing prop names or values
- ✅ Copy proven patterns from real examples

**Files to Check** (in order of priority):
1. `ComponentName.stories.tsx` - Storybook stories (PRIMARY SOURCE)
2. Other components using it - Real codebase usage
3. `ComponentName.tsx` - Component source (for prop verification)

### Component Props Validation

**After checking stories**, verify component prop types before using them. Do not guess or assume prop values exist.

**Pattern to Follow**:
1. **Check .stories.tsx file** for usage examples (FIRST!)
2. **Read the component file** to verify available props
3. **Check TypeScript interfaces** for valid prop values (enums, unions, etc.)
4. **Use only documented props** - don't invent prop names or values

**Common Mistakes to Avoid**:

```typescript
// ❌ BAD - Guessing prop values without checking
<Text variant="caption" weight="medium">Label</Text>
// "caption" and "medium" don't exist in Text component!

// ✅ GOOD - Verified valid props from Text.tsx
<Text variant="label" weight="strong">Label</Text>
// Valid variants: 'h1' | 'h2' | 'h3' | 'base' | 'body' | 'subtext' | 'label'
// Valid weights: 'normal' | 'strong' | 'stronger'

// ❌ BAD - Inventing a prop that doesn't exist
<Button size="large">Click me</Button>
// Button component doesn't have a "size" prop!

// ✅ GOOD - Using actual Button props
<Button variant="primary">Click me</Button>
```

**How to Verify Props**:
```typescript
// 1. Read the component file
// services/dashboard-ui/src/components/common/Text.tsx

// 2. Find the interface/type definition
export interface IText extends HTMLAttributes<HTMLSpanElement> {
  family?: TTextFamily
  variant?: TTextVariant  // ← Check this type
  weight?: TTextWeight    // ← Check this type
}

// 3. Find the valid values
export type TTextVariant =
  | 'h1' | 'h2' | 'h3'
  | 'base' | 'body'
  | 'subtext' | 'label'  // ← These are the ONLY valid values

export type TTextWeight = 'normal' | 'strong' | 'stronger'  // ← Only these!
```

**Best Practice**: When using any component for the first time, follow this process:
1. **Check for `.stories.tsx` file FIRST** - See real usage examples
2. **Read the component source** - Verify prop types and interfaces
3. **Look at other usage** - Search codebase for similar patterns

This discovery process prevents TypeScript errors, incorrect usage patterns, and wasted time guessing APIs.

### Key Scripts
- `npm run dev` - Development server with Turbo
- `npm run generate-api-types` - Generate types from OpenAPI spec
- `npm run generate-api-mocks` - Generate MSW mocks from API spec
- `npm run test` - Run tests with Vitest
- `npm run lint` - ESLint validation (must show zero errors)

### Code Comments - Write Less, Not More

**CRITICAL**: Avoid writing unnecessary comments. Code should be self-documenting through clear naming and structure.

**When NOT to Write Comments** (most common cases):

```typescript
// ❌ BAD - Stating the obvious
// Track success
trackEvent({ event: 'action_run', status: 'ok' })

// Close modal
removeModal(modalId)

// Get form data
const formData = new FormData(form)

// Loop through items
items.forEach(item => { ... })

// ✅ GOOD - No comments needed, code is clear
trackEvent({ event: 'action_run', status: 'ok' })
removeModal(modalId)
const formData = new FormData(form)
items.forEach(item => { ... })
```

**JSDoc Comments - Rarely Needed**:

```typescript
// ❌ BAD - Obvious from function name
/**
 * Button component that opens the InstallActionManualRunModal
 * Use this to trigger manual action workflow runs
 */
export const InstallActionManualRunButton = ({ ... }) => { ... }

// ✅ GOOD - No JSDoc needed, name is self-explanatory
export const InstallActionManualRunButton = ({ ... }) => { ... }

// ❌ BAD - Obvious from function name
/**
 * Normalize environment variables from action config steps
 * Merges env vars from all steps, with first occurrence taking precedence
 */
function normalizeEnvVars(steps) { ... }

// ✅ GOOD - No comment needed, function name describes what it does
function normalizeEnvVars(steps) { ... }
```

**When Comments ARE Useful** (rare cases):

```typescript
// ✅ GOOD - Explains non-obvious business logic
// Use double-equals to match both null and undefined from form data
if (overwrite[key] == envVars[key]) return acc

// ✅ GOOD - Warns about important constraint
// Don't use Array.at() here - not supported in Safari < 15.4
const firstItem = items[0]

// ✅ GOOD - Explains workaround for obscure bug
// Force re-render to fix React 18 hydration issue with SSR
useEffect(() => setMounted(true), [])

// ✅ GOOD - Documents complex algorithm
// Boyer-Moore string search for O(n/m) performance
function search(text, pattern) { ... }
```

**Best Practices**:
- **Default to NO comments** - Let the code speak for itself
- **Use clear naming** instead of comments to explain intent
- **Extract functions** with descriptive names instead of commenting code blocks
- **Remove obvious comments** like "// Track success", "// Close modal", "// Get data"
- **Avoid JSDoc** on simple/obvious functions and components
- **Only comment** when explaining WHY, not WHAT (and only if the "why" isn't obvious)

**Examples of Self-Documenting Code**:

```typescript
// ❌ BAD - Comment explains what code does
// Filter out unchanged variables before sending to API
const changedVars = allVars.filter(v => v.value !== v.originalValue)

// ✅ GOOD - Function name explains what it does
const changedVars = filterUnchangedVariables(allVars)

// ❌ BAD - Comment describes the logic
// Custom env vars are prefixed with "custom:" and we need to extract the name/value pairs
if (key.startsWith('custom:')) { ... }

// ✅ GOOD - Function name describes the logic
if (isCustomEnvVar(key)) {
  const { name, value } = parseCustomEnvVar(key, formData)
  ...
}
```

**Remember**: If you feel the need to write a comment, first ask: "Can I make the code clearer instead?" The answer is usually yes.

## React Server Components (RSC) Architecture

### RSC vs Client Component Patterns

**RSC Components**:
- **Location**: Live in `/src/app/*` directories next to `page.tsx`/`layout.tsx` that use them
- **File Naming**: kebab-case (e.g., `runner-health.tsx`, `install-details.tsx`)
- **Structure**: Async function components that perform server-side data fetching
- **Export Pattern**: Export both main component and error component from same file

**Client Components**:
- **Location**: Live in `/src/components/*` directories (traditional component structure)
- **Directive**: Must have `'use client'` at top of file
- **Purpose**: Handle interactivity, browser APIs, and real-time data updates
- **Data Fetching**: Use `useQuery` or `usePolling` hooks

### RSC Component Pattern

```typescript
// RSC in /src/app/[org-id]/installs/[install-id]/runner-health.tsx
import { getRunnerRecentHealthChecks } from '@/lib'

export async function RunnerHealth({
  orgId,
  runnerId,
}: {
  orgId: string
  runnerId: string
}) {
  const { data: healthchecks, error } = await getRunnerRecentHealthChecks({
    orgId,
    runnerId,
  })

  return !error ? (
    <RunnerHealthCard
      initHealthchecks={healthchecks}
      runnerId={runnerId}
      shouldPoll
    />
  ) : (
    <RunnerHealthError />
  )
}

export const RunnerHealthError = () => (
  <Card>
    <EmptyState
      emptyMessage="Runner health checks will display here once available."
      emptyTitle="No health check data"
      variant="diagram"
    />
  </Card>
)
```

### Client Component Data Fetching

**Use `useQuery` for**: One-time data fetching, user-triggered refetches
**Use `usePolling` for**: Real-time data, automatic updates with exponential backoff

**Important**: Client-side hooks (`useQuery`/`usePolling`) call **Next.js API endpoints**, not the ctl-api directly. These endpoints proxy to the same `/lib/ctl-api/` functions.

```typescript
// Client Component in /src/components/runners/RunnerHealthCard.tsx
'use client'

export const RunnerHealthCard = ({
  initHealthchecks,
  runnerId,
  shouldPoll = false,
  pollInterval = 60000,
}: {
  initHealthchecks: TRunnerHealthCheck[]
  runnerId: string
  shouldPoll?: boolean
  pollInterval?: number
}) => {
  const { org } = useOrg()
  
  // usePolling calls Next.js API endpoint (not ctl-api directly)
  const { data: healthchecks, error } = usePolling<TRunnerHealthCheck[]>({
    path: `/api/orgs/${org?.id}/runners/${runnerId}/health-checks`, // Next.js API route
    shouldPoll,
    initData: initHealthchecks, // From RSC server-side fetch
    pollInterval,
  })

  return (
    <Card>
      {/* Render with real-time data */}
    </Card>
  )
}
```

### Next.js API Endpoints (`/src/app/api/`)

Client-side data fetching requires corresponding Next.js API endpoints that proxy to `/lib/ctl-api/` functions:

```typescript
// /src/app/api/orgs/[orgId]/runners/[runnerId]/health-checks/route.ts
import { type NextRequest, NextResponse } from "next/server"
import { getRunnerRecentHealthChecks } from "@/lib"
import type { TRouteProps } from "@/types"

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<"orgId" | "runnerId">,
) {
  const { runnerId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get("limit") || undefined
  const offset = searchParams.get("offset") || undefined
  
  // Same lib/ctl-api function used by RSC components
  const response = await getRunnerRecentHealthChecks({ 
    runnerId, 
    orgId, 
    limit, 
    offset 
  })
  return NextResponse.json(response)
}
```

### API Endpoint Patterns

**Simple GET Endpoint**:
```typescript
// /src/app/api/orgs/[orgId]/installs/[installId]/route.ts
export async function GET(_: NextRequest, { params }: TRouteProps<'orgId' | 'installId'>) {
  const { installId, orgId } = await params
  const response = await getInstall({ installId, orgId })
  return NextResponse.json(response)
}
```

**GET with Query Parameters**:
```typescript
// /src/app/api/orgs/[orgId]/installs/[installId]/components/[componentId]/deploys/route.ts
export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId' | 'componentId'>
) {
  const { installId, componentId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined

  const response = await getComponentDeploys({
    installId,
    componentId,
    orgId,
    limit,
    offset,
  })
  return NextResponse.json(response)
}
```

### Data Flow Architecture

```
Client Hook → Next.js API Route → /lib/ctl-api/ Function → ctl-api Service
     ↓                ↓                    ↓                     ↓
usePolling    /api/orgs/.../route.ts   getRunner()       HTTP to ctl-api
```

### When to Create API Endpoints

**You need a Next.js API endpoint when**:
- Client components need to fetch data with `useQuery`/`usePolling`
- Real-time polling is required (health checks, job status, etc.)
- User interactions trigger data fetches (search, pagination, filtering)

**You don't need API endpoints when**:
- RSC components fetch data server-side (use `/lib/ctl-api/` directly)
- Server actions handle mutations (use server actions pattern)

### API Endpoint Organization

API routes mirror the resource hierarchy:
```
/api/orgs/[orgId]/                           # Organization context
├── runners/[runnerId]/health-checks/        # Runner health data
├── installs/[installId]/components/         # Install components
└── workflows/[workflowId]/steps/           # Workflow steps
```

This matches the `/lib/ctl-api/` function organization and the ctl-api service endpoints.

### AsyncBoundary Loading Pattern

**Purpose**: Wraps RSC components with ErrorBoundary + Suspense for granular loading states

```typescript
// Usage in page.tsx
import { AsyncBoundary } from '@/components/common/AsyncBoundary'

<AsyncBoundary
  errorFallback={<RunnerHealthError />}
  loadingFallback={<RunnerHealthCardSkeleton />}
>
  <RunnerHealth orgId={orgId} runnerId={runnerId} />
</AsyncBoundary>
```

### Data Flow Architecture

```
RSC (Server) → AsyncBoundary → Client Component
     ↓              ↓              ↓
Server fetch → Loading state → initData prop + real-time polling
```

**Key Benefits**:
- **Initial Speed**: RSC provides immediate server-rendered data
- **Real-time Updates**: Client components handle live data with polling
- **Error Isolation**: AsyncBoundary provides granular error boundaries
- **Loading States**: Custom skeletons for each component section

### Hook Selection Guide

```typescript
// ✅ useQuery - Static data, manual refresh
const { data, error, isLoading } = useQuery({
  path: '/api/endpoint',
  dependencies: [userId], // Re-fetch when userId changes
  enabled: !!userId,      // Conditional fetching
})

// ✅ usePolling - Real-time data, automatic updates
const { data, error, stopPolling } = usePolling({
  path: '/api/live-endpoint',
  shouldPoll: isActive,
  pollInterval: 30000,    // 30 seconds
  initData: initialData,  // From RSC
  backoff: {              // Exponential backoff on errors
    enabled: true,
    initialDelay: 1000,
    maxDelay: 60000,
  },
})
```

## Server-Side Data Fetching (`/lib/ctl-api/`)

### Purpose & Architecture

The `/lib/ctl-api/` functions provide **server-side data fetching** for RSC components. They directly call the Nuon ctl-api service using the shared `/lib/api.ts` function with proper authentication and organization context.

### Directory Structure

```
/src/lib/ctl-api/
├── accounts/           # Account management
├── apps/              # Applications and configs
├── installs/          # Installation management
├── runners/           # Runner operations
├── workflows/         # Workflow approvals
├── orgs/              # Organization operations
└── general/           # Health, version endpoints
```

### Function Patterns

**GET Functions** - Retrieve data:
```typescript
// Pattern: get-{resource}.ts
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
```

**Mutation Functions** - Create/Update/Delete:
```typescript
// Pattern: verb-{resource}.ts  
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

### Key Characteristics

- **Server-Only**: These functions run only on the server (RSC context)
- **Authentication**: Automatically handled via `auth0.getSession()` in `/lib/api.ts`
- **Organization Context**: Most functions require `orgId` parameter
- **Type Safety**: Return types match OpenAPI-generated TypeScript types
- **Error Handling**: Return `{ data, error, status, headers }` response format

### Testing Patterns

**Test Structure**:
```typescript
import '@test/mock-auth'           // Mock Auth0 authentication
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunner } from './get-runner'

describe('getRunner should handle response status codes', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'
  
  test('200 status', async () => {
    const { data: runner } = await getRunner({ runnerId, orgId })
    expect(runner).toHaveProperty('id')
    expect(runner).toHaveProperty('created_at')
  })

  // Test error responses with snapshots
  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getRunner({ runnerId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
```

**Test Patterns**:
- **Mock Auth**: `import '@test/mock-auth'` provides authenticated context
- **Snapshot Testing**: Error responses captured in `__snapshots__/` directories
- **Status Code Coverage**: Tests both success (200) and error responses (400, 401, 403, 404, 500)
- **Property Validation**: Verify expected properties exist on success responses

### Usage in RSC Components

```typescript
// RSC Component
import { getRunner, getRunnerSettings } from '@/lib'

export async function RunnerDetails({ orgId, runnerId }) {
  // Server-side data fetching with parallel requests
  const [{ data: runner }, { data: settings }] = await Promise.all([
    getRunner({ orgId, runnerId }),
    getRunnerSettings({ orgId, runnerId }),
  ])

  return (
    <RunnerDetailsCard 
      runner={runner} 
      settings={settings} 
    />
  )
}
```

### Common Patterns

**Parallel Data Fetching**:
```typescript
const [{ data: install }, { data: org }] = await Promise.all([
  getInstall({ installId, orgId }),
  getOrg({ orgId }),
])
```

**Error Handling**:
```typescript
const { data: runner, error } = await getRunner({ orgId, runnerId })

if (error) {
  notFound() // Next.js 404 page
}
```

**Resource Hierarchies**:
```typescript
// Common pattern: nested resource paths
/installs/{installId}/components/{componentId}/deploys
/runners/{runnerId}/jobs/{jobId}/plan
/apps/{appId}/configs/{configId}/graph
```

### Global vs Organization-Scoped Functions

- **Organization-Scoped** (most functions): Include `orgId` parameter
- **Global/Account-Level**: Functions in `/accounts/` typically don't require `orgId`

### Index File Patterns

Each directory exports its functions via `index.ts`:
```typescript
// /lib/ctl-api/runners/index.ts
export * from './get-runner'
export * from './get-runner-settings'
export * from './cancel-runner-job'
// ... other exports
```

This enables clean imports: `import { getRunner, getRunnerSettings } from '@/lib'`

## Server Actions (`/src/actions/`)

### Purpose & Architecture

Server actions in `/src/actions/` provide **server-side mutations** for form submissions, user interactions, and data modifications. They use the same `/lib/ctl-api/` functions but wrap them in the `executeServerAction()` helper for consistent behavior and automatic cache revalidation.

### Server Action Patterns

**Basic Server Action Pattern**:
```typescript
'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { buildComponent as build } from '@/lib'

export async function buildComponent({
  path,
  ...args
}: {
  componentId: string
} & IServerAction) {
  return executeServerAction({
    action: build,
    args,
    path,
  })
}
```

**Server Action with Body/Payload**:
```typescript
'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  createInstallConfig as create,
  type TCreateInstallConfigBody,
} from '@/lib'

export async function createInstallConfig({
  path,
  ...args
}: {
  body: TCreateInstallConfigBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: create,
    args,
    path,
  })
}
```

**Server Action with Custom Logic**:
```typescript
'use server'

import { auth0 } from '@/lib/auth'
import { executeServerAction } from '@/actions/execute-server-action'
import { createOrg as create, type TCreateOrgBody } from '@/lib'

export async function createOrg({
  body,
  path,
}: {
  body: TCreateOrgBody & {
    companyName?: string
    jobTitle?: string
  }
  path?: string
}) {
  const session = await auth0.getSession()

  // Custom business logic before the main action
  if (SALESFORCE_ENDPOINT) {
    await fetch(SALESFORCE_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        firstName: session?.user?.given_name,
        email: session?.user?.email,
        companyName: body?.companyName,
      }),
    }).catch((err) => {
      console.error('salesforce error:', err)
    })
  }

  return executeServerAction({
    action: create,
    args: {
      body: { name: body?.name, use_sandbox_mode: body?.use_sandbox_mode },
    },
    path,
  })
}
```

**Global/Account-Level Action** (no orgId required):
```typescript
'use server'

import { executeServerAction } from '@/actions/execute-server-action'
import { completeUserJourney as complete } from '@/lib'

export async function completeUserJourney({
  journeyName,
  path,
}: {
  journeyName: string
  path?: string
}) {
  return executeServerAction({
    action: complete,
    args: { journeyName },
    path,
  })
}
```

### executeServerAction Wrapper

The `executeServerAction` wrapper provides:
- **Automatic cache revalidation** via `revalidatePath(path)`
- **Consistent error handling** across all server actions
- **Type safety** with generic types

```typescript
// /src/actions/execute-server-action.ts
export interface IServerAction {
  orgId: string    // Most actions require organization context
  path?: string    // Path to revalidate after successful action
}

export async function executeServerAction<TArgs, TResult>({
  action,
  args,
  path,
}: {
  action: TServerActionFn<TArgs, TResult>
  args: TArgs
  path?: string
}): Promise<TResult> {
  const result = await action(args)
  if (path) revalidatePath(path)  // Invalidate Next.js cache
  return result
}
```

### Key Characteristics

- **'use server' directive**: Required at top of every server action file
- **Server-only execution**: Run in server context with full Node.js APIs
- **Same functions**: Use identical `/lib/ctl-api/` functions as RSC components
- **Cache invalidation**: Automatic `revalidatePath()` after successful operations
- **Type safety**: Inherit types from corresponding `/lib/ctl-api/` functions

### Directory Organization

Server actions mirror the resource hierarchy:
```
/src/actions/
├── accounts/       # Account-level actions (no orgId)
├── apps/          # Application actions
├── installs/      # Installation actions
├── orgs/          # Organization actions
├── runners/       # Runner management actions
└── workflows/     # Workflow approval actions
```

### Usage in Components

**Form Actions**:
```typescript
// In a form component
import { createInstallConfig } from '@/actions/installs/create-install-config'

export function InstallConfigForm({ installId, orgId }: Props) {
  const handleSubmit = async (formData: FormData) => {
    const body = {
      approval_option: formData.get('approval') as 'approve-all' | 'prompt'
    }
    
    await createInstallConfig({
      body,
      installId,
      orgId,
      path: `/orgs/${orgId}/installs/${installId}`, // Revalidate install page
    })
  }

  return (
    <form action={handleSubmit}>
      {/* Form fields */}
    </form>
  )
}
```

**Button Actions**:
```typescript
// In a button component
import { buildComponent } from '@/actions/apps/build-component'

export function BuildButton({ componentId, orgId }: Props) {
  const handleBuild = async () => {
    await buildComponent({
      componentId,
      orgId,
      path: `/orgs/${orgId}/components/${componentId}`, // Revalidate component page
    })
  }

  return <button onClick={handleBuild}>Build Component</button>
}
```

### Error Handling

Server actions inherit error handling from `/lib/ctl-api/` functions:
```typescript
// Server action returns same response format
const { data, error, status } = await createInstallConfig({
  body,
  installId,
  orgId,
  path,
})

if (error) {
  // Handle error in component
  console.error('Failed to create install config:', error)
}
```

### Path Revalidation Patterns

Common path revalidation patterns:
```typescript
// Revalidate specific resource page
path: `/orgs/${orgId}/installs/${installId}`

// Revalidate list page after creation
path: `/orgs/${orgId}/installs`

// Revalidate multiple paths (use array in newer Next.js)
path: [`/orgs/${orgId}`, `/orgs/${orgId}/apps/${appId}`]
```

### Writing New Server Actions

1. **Create file in appropriate `/src/actions/` subdirectory**
2. **Add 'use server' directive** at the top
3. **Import executeServerAction and corresponding lib function**
4. **Define function signature** matching the lib function + IServerAction
5. **Use executeServerAction wrapper** with path revalidation
6. **Export function** for use in components

```typescript
// Template for new server action
'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { libFunction as lib, type TBodyType } from '@/lib'

export async function actionName({
  path,
  ...args
}: {
  // Parameters matching lib function
  param1: string
  body?: TBodyType
} & IServerAction) {
  return executeServerAction({
    action: lib,
    args,
    path,
  })
}
```

## User Interactions with Server Actions

### Pattern Overview

User interactions (buttons, forms, modals) that trigger server actions follow a consistent pattern using `useServerAction` hook and `useServerActionToast` for user feedback. This provides loading states, error handling, and success notifications.

### Complete User Interaction Pattern

**Example: Cancel Workflow with Confirmation Modal**

```typescript
'use client'

import { usePathname } from 'next/navigation'
import { cancelWorkflow } from '@/actions/workflows/cancel-workflow'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import type { TWorkflow } from '@/types'

// Modal Component with Server Action
export const CancelWorkflowModal = ({
  workflow,
  ...props
}: { workflow: TWorkflow } & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  
  // useServerAction hook handles loading, error states, and execution
  const { data, error, isLoading, execute } = useServerAction({
    action: cancelWorkflow,
  })

  // useServerActionToast provides automatic toast notifications
  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>There was an error cancelling {workflow.type} workflow.</Text>
        <Text>{error?.error || 'Unknown error occurred.'}</Text>
      </>
    ),
    errorHeading: `${workflow.name} was not cancelled.`,
    onSuccess: () => {
      removeModal(props.modalId) // Close modal on success
    },
    successContent: <Text>Cancelled the {workflow.type} workflow.</Text>,
    successHeading: `${workflow.name} was cancelled.`,
  })

  return (
    <Modal
      heading={`Cancel ${workflow?.type} workflow?`}
      primaryActionTrigger={{
        children: isLoading ? 'Canceling workflow...' : 'Cancel workflow',
        disabled: isLoading,
        onClick: () => {
          execute({ 
            orgId: org.id, 
            path,  // For cache revalidation
            workflowId: workflow.id 
          })
        },
        variant: 'primary',
      }}
      {...props}
    >
      <Text variant="base">
        Are you sure? Once cancelled, you cannot restart this workflow.
      </Text>
    </Modal>
  )
}

// Button Component that Triggers Modal
export const CancelWorkflowButton = ({
  workflow,
  ...props
}: { workflow: TWorkflow } & IButtonAsButton) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      variant="danger"
      onClick={() => {
        addModal(<CancelWorkflowModal workflow={workflow} />)
      }}
      {...props}
    >
      Cancel workflow
    </Button>
  )
}
```

### useServerAction Hook

**Purpose**: The primary way to call a server action from a client component and use its output.

```typescript
const { data, error, isLoading, execute } = useServerAction({
  action: cancelWorkflow, // The server action function
})

// Call execute() to invoke the server action
await execute({
  orgId: org.id,
  path: usePathname(), // For cache revalidation
  workflowId: workflow.id
})
```

**What it provides**:
- **`execute()` function**: Call this to invoke your server action with parameters
- **`data`**: The successful response data from the server action
- **`error`**: Error object if the server action failed
- **`isLoading`**: Boolean indicating if the action is currently executing
- **`status` & `headers`**: Additional response metadata

**Key Point**: `useServerAction` is the standard way to call server actions from client components and access their results.

### useServerActionToast Hook

**Purpose**: Automatically shows success or error toast notifications after a server action resolves.

```typescript
useServerActionToast({
  data,        // Pass data from useServerAction
  error,       // Pass error from useServerAction
  successHeading: 'Workflow Cancelled!',
  successContent: <Text>The workflow has been cancelled successfully.</Text>,
  errorHeading: 'Cancellation Failed',
  errorContent: <Text>{error?.error || 'Something went wrong.'}</Text>,
  onSuccess: () => {
    // Custom success handler (close modal, redirect, etc.)
    removeModal(modalId)
  },
  onError: () => {
    // Custom error handler
    console.error('Action failed:', error)
  },
})
```

**What it does**:
- **Watches `data` and `error`** from `useServerAction`
- **Shows success toast** when `data` exists and no error
- **Shows error toast** when `error` exists and no data
- **Calls `onSuccess()`** after showing success toast
- **Calls `onError()`** after showing error toast

**Key Point**: Use `useServerActionToast` to provide automatic user feedback after server actions complete, eliminating the need to manually manage toast notifications.

### Common Interaction Patterns

**1. Simple Button Action** (no confirmation):
```typescript
export function BuildButton({ componentId }: Props) {
  const { org } = useOrg()
  const path = usePathname()
  const { data, error, isLoading, execute } = useServerAction({
    action: buildComponent,
  })

  useServerActionToast({
    data,
    error,
    successHeading: 'Build Started',
    errorHeading: 'Build Failed',
  })

  return (
    <Button
      disabled={isLoading}
      onClick={() => execute({ componentId, orgId: org.id, path })}
    >
      {isLoading ? 'Building...' : 'Build Component'}
    </Button>
  )
}
```

**2. Form Submission**:
```typescript
export function CreateInstallForm({ appId }: Props) {
  const { org } = useOrg()
  const path = usePathname()
  const { data, error, isLoading, execute } = useServerAction({
    action: createAppInstall,
  })

  useServerActionToast({
    data,
    error,
    successHeading: 'Install Created',
    onSuccess: () => router.push(`/${org.id}/installs/${data.id}`),
  })

  const handleSubmit = async (formData: FormData) => {
    const body = {
      name: formData.get('name') as string,
      region: formData.get('region') as string,
    }

    await execute({ appId, body, orgId: org.id, path })
  }

  return (
    <form action={handleSubmit}>
      <input name="name" required />
      <select name="region" required>
        <option value="us-east-1">US East 1</option>
      </select>
      <Button type="submit" disabled={isLoading}>
        {isLoading ? 'Creating...' : 'Create Install'}
      </Button>
    </form>
  )
}
```

**3. Modal Confirmation** (as shown in full example above):
- Button triggers modal
- Modal contains form/confirmation
- Server action executed on confirmation
- Modal closes on success

### Loading States & Disabled UI

```typescript
// Button loading state
<Button disabled={isLoading}>
  {isLoading ? (
    <>
      <Icon variant="Loading" />
      Processing...
    </>
  ) : (
    'Submit'
  )}
</Button>

// Form field disabled during loading
<input disabled={isLoading} />

// Modal primary action disabled
primaryActionTrigger={{
  disabled: isLoading,
  onClick: () => execute({...params}),
}}
```

### Error Handling in UI

```typescript
// Inline error banner in modal/form
{error && (
  <Banner theme="error">
    {error?.error || 'An error occurred, please try again.'}
  </Banner>
)}

// Custom error content in toast
errorContent: (
  <>
    <Text>Operation failed for {resource.name}</Text>
    <Text variant="caption">{error?.description}</Text>
  </>
)
```

### Complete Flow Summary

1. **User Interaction** → Button click or form submission
2. **State Management** → `useServerAction` provides loading/error states  
3. **Action Execution** → `execute()` calls server action with parameters
4. **User Feedback** → `useServerActionToast` shows success/error notifications
5. **UI Updates** → Cache revalidation refreshes data, modals close, navigation occurs

This pattern ensures consistent UX across all server action interactions with proper loading states, error handling, and user feedback.

## TypeScript Conventions

### Naming Patterns

**Types**: Use `T` prefix for all data types and API response types
```typescript
// ✅ Good - Type naming
type TApp = {
  id: string
  name: string
  created_at: string
}

type TInstall = {
  id: string
  app_id: string
  status: 'pending' | 'deployed' | 'failed'
}

type TWorkflow = {
  id: string
  type: 'deploy' | 'destroy'
  status: string
}
```

**Interfaces**: Use `I` prefix for component props and configuration interfaces
```typescript
// ✅ Good - Interface naming
interface ICard {
  children: ReactNode
  className?: string
  variant?: 'default' | 'outlined'
}

interface IDropdown {
  options: Array<{ label: string; value: string }>
  value?: string
  onChange: (value: string) => void
  disabled?: boolean
}

interface IModal {
  isOpen: boolean
  onClose: () => void
  heading?: string
  children: ReactNode
}
```

### Usage Examples
```typescript
// API response types (T prefix)
const { data: app }: { data: TApp } = await getApp({ appId, orgId })
const { data: installs }: { data: TInstall[] } = await getInstalls({ orgId })

// Component props (I prefix)
export const Card = ({ children, className, variant }: ICard) => {
  // Component implementation
}

export const AppCard = ({ app }: { app: TApp }) => {
  // Component that uses TApp type
}
```

### Type Generation & Import Pattern

**API Types** (`T` prefix): Auto-generated from ctl-api OpenAPI spec but require manual type extraction.

**CRITICAL PATTERN**: Never import types directly from the generated file. Always use the `/src/types/ctl-api.types.ts` intermediary file.

```typescript
// ❌ NEVER - Direct import from generated file
import type { components } from '@/types/nuon-oapi-v3'
type TApp = components['schemas']['app.App']  // Don't do this in components

// ✅ ALWAYS - Import from ctl-api.types.ts
import type { TApp, TInstall, TWorkflow } from '@/types/ctl-api.types'
```

**Type Extraction Process**:
1. OpenAPI spec generates complex nested types in `/src/types/nuon-oapi-v3.ts`
2. `/src/types/ctl-api.types.ts` extracts and renames them with `T` prefix
3. Components and lib functions import from `ctl-api.types.ts`

**How to Add New Types**:
```typescript
// In /src/types/ctl-api.types.ts
import { components } from '@/types/nuon-oapi-v3'

// Extract and rename with T prefix
export type TNewResource = components['schemas']['app.NewResource']

// Enhance with additional properties if needed
export type TEnhancedInstall = components['schemas']['app.Install'] & {
  app?: components['schemas']['app.App']
  org_id?: string
}
```

**Component Interfaces** (`I` prefix): Manually defined in component files or shared in `/src/types/`

## Important Notes

- **Ignore `/old/` directories**: Contains deprecated components and utilities
- **Mock Service Worker**: Tests automatically use MSW instead of real API calls
- **Type Safety**: TypeScript types auto-generated from ctl-api OpenAPI spec
- **Path Revalidation**: Server actions use `executeServerAction()` for automatic cache invalidation
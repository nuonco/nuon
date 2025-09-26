# Dashboard UI Service  

The **Dashboard UI** is the primary web application frontend for the Nuon platform, providing a comprehensive interface for managing applications, deployments, and infrastructure.

## Service Overview

This is the main user-facing web application built with Next.js and React. It provides a complete dashboard experience for developers and operators to manage their BYOC (Bring Your Own Cloud) deployments through the Nuon platform.

## Architecture

- **Framework**: Next.js 15+ with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS with custom design system
- **State Management**: React hooks with API data fetching
- **Authentication**: Auth0 integration via `@auth0/nextjs-auth0`
- **Testing**: Vitest with React Testing Library
- **Build Tool**: Turbo for development speed

## Authentication Patterns

### Server Actions vs API Routes for Authenticated Operations

**IMPORTANT**: For authenticated operations that need to interact with the Nuon API, use **server actions** instead of API routes.

**Server Actions (Recommended)**:
- Run in the server component context where Auth0 session access works properly
- Can directly use `getFetchOpts()` for authenticated API calls
- Follow the same pattern as existing mutations in `/src/components/*-actions.ts`
- Example: `createOrganization()` in `org-actions.ts`

**API Routes (Avoid for Auth)**:
- Run in a different context where Auth0 session may not be properly accessible
- Can cause 401 authentication errors even with valid user sessions
- Should primarily be used for public endpoints or proxy functionality

**Best Practice**: 
```typescript
// ✅ Good - Server action pattern
'use server'
export async function createSomething(data: any) {
  return fetch(`${API_URL}/v1/endpoint`, {
    ...(await getFetchOpts()),
    method: 'POST',
    body: JSON.stringify(data),
  })
}

// ❌ Avoid - API route for authenticated operations  
export const POST = async (request: NextRequest) => {
  // Auth0 session access issues can occur here
}
```

## Relationship to Other Services

- **Primary API**: Consumes `ctl-api` service for all backend operations
- **Authentication**: Integrates with Auth0 for user authentication
- **Monitoring**: DataDog RUM and logging integration
- **Analytics**: Segment analytics for user behavior tracking
- **Infrastructure**: Deployed to Kubernetes via Helm charts

## Project Structure

### Core Files
- `next.config.mjs` - Next.js configuration
- `tailwind.config.ts` - Tailwind CSS customization
- `package.json` - Dependencies and scripts
- `Dockerfile` - Container build configuration

### Key Directories

#### `/src/app/` - Next.js App Router
Modern Next.js file-based routing structure:
- `[org-id]/` - Organization-scoped routes
  - `apps/[app-id]/` - Application management pages
  - `installs/[install-id]/` - Installation management and monitoring
  - `releases/` - Release management interface
  - `runner/` - Runner status and management
  - `team/` - Organization team management
- `api/` - API route handlers and proxy endpoints
- `stratus/` - New UI framework implementation

#### `/src/components/` - React Components
Comprehensive component library organized by domain:
- `Apps/` - Application-specific components
- `Components/` - Build and component management
- `Installs/` - Installation management UI
- `InstallComponents/` - Deployment component management
- `InstallWorkflows/` - Workflow and approval interfaces
- `Runners/` - Runner monitoring and management
- `Orgs/` - Organization administration
- `LogStream/` - Real-time log viewing
- Common UI components: `Button`, `Modal`, `DataTable`, etc.

#### `/src/stratus/` - New Design System
Modern React component library:
- `components/` - Reusable UI components
- `context/` - React context providers
- `actions/` - Server actions for data mutations

#### `/src/lib/` - Business Logic
API client libraries and utilities:
- `apps.ts` - Application management logic
- `installs.ts` - Installation operations
- `components.ts` - Component build management
- `orgs.ts` - Organization operations
- `runners.ts` - Runner communication

#### `/src/utils/` - Utility Functions
- `auth.ts` - Authentication helpers
- `query-data.ts` - API data fetching
- `mutate-data.ts` - Data mutation helpers
- `time-utils.ts` - Date/time formatting
- `datadog-*.tsx` - Monitoring integration

#### `/src/types/` - TypeScript Definitions
- `nuon-oapi-v3.d.ts` - Auto-generated API types from OpenAPI spec
- `ctl-api.types.ts` - Custom API type definitions
- `dashboard.types.ts` - Dashboard-specific types

### Infrastructure

#### `/infra/` - Terraform Configuration
- `service.tf` - ECS/EKS service definition
- `certificate.tf` - SSL certificate management
- Infrastructure deployment configuration

#### `/k8s/` - Kubernetes Deployment
Helm chart for Kubernetes deployment:
- `templates/` - Kubernetes resource templates
- `values.yaml` - Configuration values

## Key Features

### Multi-Organization Support
- Organization switching and context management
- Role-based access control and permissions
- Team member management and invitations

### Application Management
- App creation and configuration
- Component dependency management
- Build history and status tracking
- Configuration templating and validation

### Installation Management
- Infrastructure provisioning workflows
- Component deployment orchestration
- Real-time status monitoring
- Approval workflows for sensitive operations

### Workflow Management
- Visual workflow execution tracking
- Step-by-step approval processes
- Terraform plan reviewing and approval
- Deployment rollback capabilities

### Monitoring & Observability  
- Real-time log streaming
- Runner health monitoring
- Deployment status tracking
- Infrastructure state visualization

### Developer Experience
- Code editor integration (Monaco)
- Terraform/YAML syntax highlighting
- Configuration diff viewing
- Template rendering and validation

## Development

### Setup
```bash
cd services/dashboard-ui
npm install
npm run dev
```

### Key Scripts
- `npm run dev` - Development server with Turbo
- `npm run build` - Production build
- `npm run generate-api-types` - Generate types from OpenAPI spec
- `npm run test` - Run tests with Vitest
- `npm run lint` - ESLint validation

### Code Quality & Linting Rules

**IMPORTANT**: All updates to the dashboard-ui project must follow the established linting and formatting rules.

#### Configuration Files
Agents must read and follow the rules defined in these configuration files:

- **ESLint Rules**: `.eslintrc.json` - Defines code quality rules including console usage, React/JSX patterns, and import validation
- **Prettier Formatting**: `.prettierrc` - Defines code formatting standards (trailing commas, quotes, semicolons)
- **TypeScript Config**: `tsconfig.json` - TypeScript compiler and type checking rules

#### Validation Process
```bash
# Always run before committing changes
npm run lint

# Must show: "✔ No ESLint warnings or errors"
```

#### Best Practice for Agents
1. **Read config files first**: Always examine `.eslintrc.json` and `.prettierrc` to understand current rules
2. **Run lint checks**: Use `npm run lint` to identify violations before making changes
3. **Fix all violations**: Ensure zero ESLint errors/warnings before completing tasks
4. **Follow existing patterns**: Match the coding style and patterns used throughout the codebase

### API Integration
The dashboard automatically generates TypeScript types from the ctl-api OpenAPI specification, ensuring type safety across the entire application.

### Testing
- Unit tests with Vitest
- Component tests with React Testing Library
- API mocking with MSW (Mock Service Worker)

## Configuration

### Environment Variables
- `NUON_API_URL` - Backend API endpoint
- Auth0 configuration for authentication
- DataDog keys for monitoring
- Feature flags for experimental features

### Build Configuration
- Next.js App Router configuration
- Tailwind CSS with custom design tokens
- TypeScript strict mode enabled
- Bundle analysis and optimization

## Deployment

The dashboard is containerized and deployed to Kubernetes:
- Docker multi-stage build for optimization
- Helm chart deployment with environment-specific values
- SSL termination and custom domain support
- Auto-scaling based on traffic

This service provides the primary user interface for the entire Nuon platform, enabling users to manage complex cloud deployments through an intuitive web interface.

## User Journey Modal System

The dashboard implements a sophisticated modal system for guided user onboarding:

### Modal Architecture Pattern
Journey-based modals use layout-level wrapper components that conditionally render based on user journey state:

```typescript
// Layout wrapper pattern
export const PageWithModal = ({ children }) => {
  const { account } = useAccount()
  const [showModal, setShowModal] = useState(false)
  const [userDismissed, setUserDismissed] = useState(false) // CRITICAL: Prevents reopen loops
  
  useEffect(() => {
    const shouldShow = journeyConditionsMet && !userDismissed
    setShowModal(shouldShow)
  }, [account, userDismissed])
  
  const handleClose = () => {
    setShowModal(false)
    setUserDismissed(true) // Prevents immediate reopening
    await refreshAccount() // Check for journey updates
  }
}
```

### Current Modal Implementations

#### App Creation Modal
- **Location**: `src/components/Apps/AppsPageWithModal.tsx`
- **Trigger**: User has no apps and `app_created` step incomplete
- **Behavior**: Guides user to create first app
- **Navigation**: On `app_created` completion with app ID → navigate to app page

#### Install Creation Modal  
- **Location**: `src/components/Apps/AppPageWithInstallModal.tsx`
- **Trigger**: User on app page, `app_created` complete, `install_created` incomplete
- **Behavior**: Guides user to create first install
- **Navigation**: Modal hides when `install_created` step completes

### Critical Modal State Management

**Problem**: Modal reopen loops after dismissal
**Root Cause**: useEffect reopens modal after `refreshAccount()` if journey conditions still met
**Solution**: Track manual dismissal separately from journey completion

```typescript
// CRITICAL PATTERN: Always track user dismissal
const [userDismissed, setUserDismissed] = useState(false)

// Modal shows only if conditions met AND not manually dismissed
const shouldShow = journeyConditionsMet && !userDismissed

const handleClose = async () => {
  setShowModal(false)
  setUserDismissed(true) // Prevents loop
  await refreshAccount() // Safe to refresh now
}
```

### Journey Step Interface
Frontend interfaces must match backend journey structure:

```typescript
interface UserJourney {
  name: string
  title: string
  steps: Array<{
    name: string
    title: string
    complete: boolean
    app_id?: string      // For navigation to specific app
    install_id?: string  // For navigation to specific install
  }>
}
```

### Modal Component Patterns

#### Close Button Configuration
```typescript
<Modal
  isOpen={isOpen}
  onClose={onClose}
  showCloseButton={false} // Only action button for guided flow
  actions={<Button onClick={onClose}>Got it</Button>}
>
```

#### Layout Integration
```typescript
// In layout.tsx
<AppProvider>
  <PageWithModal>
    {children}
  </PageWithModal>  
</AppProvider>
```

### Journey-Based Navigation
When journey steps complete with entity IDs, trigger automatic navigation:

```typescript
// Navigation on journey completion
if (appCreatedStep?.complete && appCreatedStep?.app_id) {
  router.push(`/${orgId}/apps/${appCreatedStep.app_id}`)
}
```

### Best Practices
- **Non-blocking**: Modals are always dismissible
- **Contextual**: Only show on relevant pages
- **Progressive**: Guide users through logical flow
- **State-safe**: Prevent infinite reopen loops
- **Entity-aware**: Use stored IDs for specific navigation

## API Integration & Error Handling Patterns

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

### Error Handling Anti-Patterns

**Problem**: Silent error handling makes debugging extremely difficult:

```typescript
// ❌ ANTI-PATTERN - Silent failures
const apiCall = async (): Promise<boolean> => {
  try {
    const { error } = await api(config)
    if (error) {
      return false  // No logging, no user feedback
    }
    return true
  } catch (err) {
    return false    // Completely silent
  }
}
```

**Symptoms**:
- Buttons appear to "do nothing" when clicked
- API failures invisible to users and developers
- Authentication errors go unnoticed

**Debugging Approach**:
1. Temporarily add console.log statements (disable ESLint no-console rule)
2. Use browser DevTools → Network tab to inspect API requests
3. Check for 401/403 authentication errors
4. Verify request headers include proper authentication

### Modal State Management Complexities

**Race Condition Risk**: Modal visibility controlled by backend data can create timing issues:

```typescript
// Modal auto-reopens based on journey state
useEffect(() => {
  if (shouldShowModal()) {
    setShowModal(true)  // Could reopen immediately after API success
  }
}, [account])  // Triggers on every account refresh
```

**Problem**: Even after successful API calls, modal might reopen if:
- Backend data update has delay
- Account refresh occurs before backend sync
- Race condition between API response and state update

**Debugging**: Check timing between API success, account refresh, and modal state changes.

### Authentication Context in Client Components

**Pattern**: Client-side API calls use `auth0.getSession()` for authentication:

```typescript
// From /lib/api.ts
const session = await auth0.getSession()
const fetchOpts: RequestInit = {
  headers: {
    Authorization: `Bearer ${session?.tokenSet?.accessToken}`,
    'X-Nuon-Org-ID': orgId || '',
    // ...
  }
}
```

**Potential Issues**:
- Auth0 session may not be available in all rendering contexts
- Client component authentication differs from server action patterns
- Session token expiration can cause silent failures

### ESLint Configuration Impact

**Development Challenge**: The `no-console` ESLint rule prevents debugging console statements:

```typescript
console.log('Debug info')  // ❌ ESLint error
```

**Workarounds**:
- Temporarily disable rule with `// eslint-disable-next-line no-console`
- Use browser DevTools for runtime debugging
- Implement proper error boundaries for user-facing error handling

### Complex Modal State Dependencies

**Pattern**: UserJourneyProvider manages modal state with intricate dependencies:

```typescript
// Modal closes only when ALL journey steps complete
const allStepsComplete = journey?.steps.every(step => step.complete) ?? false
if (allStepsComplete) {
  setShowModal(false)
} else {
  await refreshAccount()  // Just refresh, don't close
}
```

**Debugging**: Modal behavior depends on:
1. API call success/failure
2. Backend data consistency
3. Account refresh timing
4. Journey step completion state
5. Modal dismissal tracking

These patterns highlight the complexity of debugging full-stack authentication and state management issues in modern React applications.
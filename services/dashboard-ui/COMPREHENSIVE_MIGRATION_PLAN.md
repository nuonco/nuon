# Dashboard UI: Comprehensive Next.js to SPA Migration Plan

## Executive Summary

This document provides an exhaustive migration plan for all 40+ pages in the dashboard-ui from Next.js App Router to React Router SPA. The analysis is based on comprehensive exploration of the Next.js app directory structure, API routes, and component patterns.

**Current Status**: 
- ✅ 5 pages have basic layout conversion (but need full functionality)
- ❌ 35+ pages still need implementation
- ⚠️ Several "completed" pages are placeholders only

**Key Finding**: Many pages marked as "migrated" only have the layout pattern converted (PageLayout → PageSection) but lack the actual functionality from the Next.js versions.

---

## Migration Patterns Reference

### Standard Pattern Conversion

**Old Next.js Pattern:**
```typescript
// app/[org-id]/page.tsx
export default async function Page({ params }) {
  const { 'org-id': orgId } = await params
  // Server-side data fetching
  return <PageLayout>...</PageLayout>
}
```

**New SPA Pattern:**
```typescript
// pages/org/OrgPage.tsx
export default function OrgPage() {
  const { orgId } = useParams()
  const { org } = useOrg()
  const { data } = usePolling({ path: '...', pollInterval: 20000 })
  return <PageSection isScrollable>...</PageSection>
}
```

### Key Differences
- `async params` → `useParams()` hook
- `async searchParams` → `useSearchParams()` hook  
- Server components → Client components with hooks
- `PageLayout` + `PageContent` → `PageSection`
- Server data fetching → `usePolling()` or `useQuery()`

---

## Page Inventory (40+ Pages)

### Organization Level (6 pages)

#### 1. Home Page / Organization Overview
- **Next.js**: `/app/[org-id]/page.tsx`
- **SPA**: `/pages/HomePage.tsx`
- **Status**: ✅ Partially migrated (needs verification)
- **Features**:
  - Recent activity feed
  - Quick stats (installs, apps, runners)
  - Getting started guide for new users
  - User journey integration
- **Components**: Dashboard cards, activity timeline
- **APIs**: 
  - `GET /api/ctl-api/v1/orgs/{orgId}/activity`
  - `GET /api/ctl-api/v1/orgs/{orgId}/stats`
- **Priority**: HIGH - Entry point for all users

#### 2. Apps List Page
- **Next.js**: `/app/[org-id]/apps/page.tsx`
- **SPA**: `/pages/org/AppsPage.tsx`
- **Status**: ⚠️ Layout converted, needs full implementation
- **Features**:
  - Apps table with search/filter
  - App status indicators
  - Quick actions (sync, configure)
  - Create new app flow
- **Components**: `AppsTable`, app creation modal
- **APIs**: 
  - `GET /api/ctl-api/v1/orgs/{orgId}/apps`
- **Priority**: HIGH - Core functionality

#### 3. Installs List Page
- **Next.js**: `/app/[org-id]/installs/page.tsx`
- **SPA**: `/pages/org/InstallsPage.tsx`
- **Status**: ⚠️ Layout converted, needs full implementation
- **Features**:
  - Installs table with filtering
  - Health status indicators
  - Quick navigation to install details
  - Create install flow
- **Components**: `InstallsTable`, install creation modal
- **APIs**: 
  - `GET /api/ctl-api/v1/orgs/{orgId}/installs`
- **Priority**: HIGH - Core functionality

#### 4. Team Management Page
- **Next.js**: `/app/[org-id]/team/page.tsx`
- **SPA**: `/pages/org/TeamPage.tsx`
- **Status**: ❌ PLACEHOLDER ONLY - User reported "shows nothing like next.js"
- **Features** (from Next.js):
  - Team members table with roles
  - Invite new members flow
  - Role assignment (Admin, Installer, Runner)
  - Member removal
  - Pending invitations management
- **Components**: Team members table, invite modal, role selector
- **APIs**: 
  - `GET /api/ctl-api/v1/orgs/{orgId}/accounts`
  - `GET /api/ctl-api/v1/orgs/{orgId}/invites`
  - `POST /api/ctl-api/v1/orgs/{orgId}/invites`
  - `DELETE /api/ctl-api/v1/orgs/{orgId}/accounts/{accountId}`
- **Priority**: CRITICAL - User explicitly requested

#### 5. Runner / Builds Page
- **Next.js**: `/app/[org-id]/runner/page.tsx`
- **SPA**: `/pages/org/OrgRunner.tsx`
- **Status**: ❌ PLACEHOLDER ONLY - User reported "shows nothing like next.js"
- **Features** (from Next.js):
  - Runner health status overview
  - Recent jobs list with status
  - Runner configuration details
  - Job queue and execution history
  - Runner logs access
  - Performance metrics
- **Components**: Runner health cards, jobs table, config panel
- **APIs**: 
  - `GET /api/ctl-api/v1/orgs/{orgId}/runner`
  - `GET /api/ctl-api/v1/orgs/{orgId}/runner/jobs`
  - `GET /api/ctl-api/v1/orgs/{orgId}/runner/health`
- **Priority**: CRITICAL - User explicitly requested

#### 6. Organization Settings Page
- **Next.js**: `/app/[org-id]/settings/page.tsx`
- **SPA**: Likely `/pages/org/OrgSettings.tsx` (needs creation)
- **Status**: ❌ Not implemented
- **Features**:
  - Organization name/details
  - Billing information
  - Feature flags
  - Danger zone (delete org)
- **Components**: Settings forms, confirmation modals
- **APIs**: 
  - `GET /api/ctl-api/v1/orgs/{orgId}`
  - `PUT /api/ctl-api/v1/orgs/{orgId}`
- **Priority**: MEDIUM

---

### App Level (8 pages)

#### 7. App Overview / Dashboard
- **Next.js**: `/app/[org-id]/apps/[app-id]/page.tsx`
- **SPA**: Likely in `/pages/apps/AppOverview.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - App metadata and status
  - Recent builds
  - Recent installs
  - Quick actions
- **Components**: App header, builds table, installs table
- **APIs**: 
  - `GET /api/ctl-api/v1/apps/{appId}`
  - `GET /api/ctl-api/v1/apps/{appId}/builds`
- **Priority**: HIGH

#### 8. App Components List
- **Next.js**: `/app/[org-id]/apps/[app-id]/components/page.tsx`
- **SPA**: `/pages/apps/AppComponents.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Components table
  - Component type indicators
  - Configuration status
  - Dependencies view
- **Components**: `ComponentsTable`
- **APIs**: 
  - `GET /api/ctl-api/v1/apps/{appId}/components`
- **Priority**: HIGH

#### 9. App Component Detail
- **Next.js**: `/app/[org-id]/apps/[app-id]/components/[component-id]/page.tsx`
- **SPA**: `/pages/apps/AppComponentDetail.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Component configuration
  - Dependencies graph
  - Version history
  - Edit configuration
- **Components**: Component config editor, dependencies graph
- **APIs**: 
  - `GET /api/ctl-api/v1/apps/{appId}/components/{componentId}`
- **Priority**: MEDIUM

#### 10. App Builds List
- **Next.js**: `/app/[org-id]/apps/[app-id]/builds/page.tsx`
- **SPA**: `/pages/apps/AppBuilds.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Builds table with pagination
  - Build status indicators
  - Trigger new build
  - Build artifacts links
- **Components**: Builds table, trigger build button
- **APIs**: 
  - `GET /api/ctl-api/v1/apps/{appId}/builds`
  - `POST /api/ctl-api/v1/apps/{appId}/builds`
- **Priority**: HIGH

#### 11. App Build Detail
- **Next.js**: `/app/[org-id]/apps/[app-id]/builds/[build-id]/page.tsx`
- **SPA**: `/pages/apps/AppBuildDetail.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Build status and timeline
  - Build logs
  - Artifacts list
  - Component build details
- **Components**: Build timeline, logs viewer, artifacts table
- **APIs**: 
  - `GET /api/ctl-api/v1/builds/{buildId}`
  - `GET /api/ctl-api/v1/builds/{buildId}/logs`
- **Priority**: MEDIUM

#### 12. App Installs (via App context)
- **Next.js**: `/app/[org-id]/apps/[app-id]/installs/page.tsx`
- **SPA**: `/pages/apps/AppInstalls.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Filtered installs table (only this app)
  - Install health status
  - Quick navigation to install details
- **Components**: `InstallsTable` (filtered)
- **APIs**: 
  - `GET /api/ctl-api/v1/apps/{appId}/installs`
- **Priority**: MEDIUM

#### 13. App Configuration / Inputs
- **Next.js**: `/app/[org-id]/apps/[app-id]/config/page.tsx`
- **SPA**: `/pages/apps/AppConfig.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - App-level config variables
  - Input definitions
  - Default values
  - Validation rules
- **Components**: Config editor, input definitions table
- **APIs**: 
  - `GET /api/ctl-api/v1/apps/{appId}/config`
  - `PUT /api/ctl-api/v1/apps/{appId}/config`
- **Priority**: MEDIUM

#### 14. App Settings
- **Next.js**: `/app/[org-id]/apps/[app-id]/settings/page.tsx`
- **SPA**: `/pages/apps/AppSettings.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - App name/description
  - VCS connection
  - Build configuration
  - Danger zone (delete app)
- **Components**: Settings forms, VCS selector
- **APIs**: 
  - `GET /api/ctl-api/v1/apps/{appId}`
  - `PUT /api/ctl-api/v1/apps/{appId}`
- **Priority**: MEDIUM

---

### Install Level (26+ pages)

#### 15. Install Overview / Dashboard
- **Next.js**: `/app/[org-id]/installs/[install-id]/page.tsx`
- **SPA**: `/pages/installs/InstallOverview.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Install metadata and status
  - Health indicators
  - Quick stats
  - Recent activity
- **Components**: Install header, health cards, activity feed
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}`
- **Priority**: COMPLETE

#### 16. Install Components List
- **Next.js**: `/app/[org-id]/installs/[install-id]/components/page.tsx`
- **SPA**: `/pages/installs/InstallComponents.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Install components table
  - Deploy status per component
  - Quick deploy actions
  - Component health
- **Components**: `InstallComponentsTable`
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/components`
- **Priority**: COMPLETE

#### 17. Install Component Detail
- **Next.js**: `/app/[org-id]/installs/[install-id]/components/[component-id]/page.tsx`
- **SPA**: `/pages/installs/InstallComponentDetail.tsx`
- **Status**: ⚠️ Basic implementation - needs enhancement
- **Current**: Shows component header and latest deploy
- **Missing**:
  - Deploy history table
  - Component logs access
  - Configuration diff viewer
  - Rollback functionality
  - Dependencies view
- **Components**: `InstallComponentHeader`, deploys table, logs viewer
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/components/{componentId}`
  - `GET /api/ctl-api/v1/installs/{installId}/components/{componentId}/deploys`
- **Priority**: MEDIUM - Enhancement needed

#### 18. Install Actions List
- **Next.js**: `/app/[org-id]/installs/[install-id]/actions/page.tsx`
- **SPA**: `/pages/installs/InstallActions.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Actions table with recent runs
  - Run status indicators
  - Trigger action button
  - Run history
- **Components**: `InstallActionsTable`
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/action-workflows`
- **Priority**: COMPLETE

#### 19. Install Action Detail
- **Next.js**: `/app/[org-id]/installs/[install-id]/actions/[action-id]/page.tsx`
- **SPA**: `/pages/installs/InstallActionDetail.tsx`
- **Status**: ⚠️ Basic implementation - needs enhancement
- **Current**: Shows action name and recent runs table
- **Missing**:
  - Action configuration display
  - Schedule information (if cron-based)
  - Success/failure statistics
  - Quick trigger button
  - Full run history pagination
- **Components**: Action header, runs table, config display
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/action-workflows/{actionId}`
  - `GET /api/ctl-api/v1/installs/{installId}/action-workflows/{actionId}/recent-runs`
- **Priority**: MEDIUM - Enhancement needed

#### 20. Install Action Run Summary (nested layout)
- **Next.js**: `/app/[org-id]/installs/[install-id]/actions/[action-id]/runs/[run-id]/page.tsx`
- **SPA**: `/pages/installs/InstallActionRunSummary.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Run status and metadata
  - Execution timeline
  - Summary statistics
  - Links to logs
- **Components**: Run timeline, status cards
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/action-workflows/{actionId}/runs/{runId}`
- **Priority**: COMPLETE

#### 21. Install Action Run Logs (nested layout)
- **Next.js**: `/app/[org-id]/installs/[install-id]/actions/[action-id]/runs/[run-id]/logs/page.tsx`
- **SPA**: `/pages/installs/InstallActionRunLogs.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Real-time log streaming
  - Log filtering
  - Download logs
  - Error highlighting
- **Components**: `UnifiedLogsProvider`, log viewer
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/action-workflows/{actionId}/runs/{runId}/logs`
- **Priority**: COMPLETE

#### 22. Install Workflows List
- **Next.js**: `/app/[org-id]/installs/[install-id]/workflows/page.tsx`
- **SPA**: `/pages/installs/InstallWorkflows.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Workflows table with status
  - Recent executions
  - Workflow type indicators
  - Quick navigation
- **Components**: Workflows table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/workflows`
- **Priority**: COMPLETE

#### 23. Install Workflow Detail
- **Next.js**: `/app/[org-id]/installs/[install-id]/workflows/[workflow-id]/page.tsx`
- **SPA**: `/pages/installs/InstallWorkflowDetail.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Workflow timeline with steps
  - Step details panel
  - Status progression
  - Logs access per step
- **Components**: `WorkflowTimeline`, `StepDetailPanel`
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/workflows/{workflowId}`
- **Priority**: COMPLETE

#### 24. Install Policies List
- **Next.js**: `/app/[org-id]/installs/[install-id]/policies/page.tsx`
- **SPA**: `/pages/installs/InstallPolicies.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Policies table
  - Evaluation status
  - Policy details
  - Pass/fail indicators
- **Components**: Policies table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/policies`
- **Priority**: COMPLETE

#### 25. Install Policy Detail
- **Next.js**: `/app/[org-id]/installs/[install-id]/policies/[policy-id]/page.tsx`
- **SPA**: `/pages/installs/InstallPolicyDetail.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Policy configuration
  - Recent evaluations
  - Compliance status
  - Evaluation history
- **Components**: Policy config display, evaluations table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/policies/{policyId}`
  - `GET /api/ctl-api/v1/installs/{installId}/policies/{policyId}/evaluations`
- **Priority**: LOW

#### 26. Install Roles List
- **Next.js**: `/app/[org-id]/installs/[install-id]/roles/page.tsx`
- **SPA**: `/pages/installs/InstallRoles.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Roles table
  - Role assignments
  - Permission details
- **Components**: Roles table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/roles`
- **Priority**: COMPLETE

#### 27. Install Role Detail
- **Next.js**: `/app/[org-id]/installs/[install-id]/roles/[role-id]/page.tsx`
- **SPA**: `/pages/installs/InstallRoleDetail.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Role permissions list
  - Account assignments
  - Edit permissions
  - Add/remove members
- **Components**: Permissions editor, members table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/roles/{roleId}`
  - `GET /api/ctl-api/v1/installs/{installId}/roles/{roleId}/accounts`
- **Priority**: LOW

#### 28. Install Stacks List
- **Next.js**: `/app/[org-id]/installs/[install-id]/stacks/page.tsx`
- **SPA**: `/pages/installs/InstallStacks.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Stacks table
  - Stack status
  - Region information
  - Quick actions
- **Components**: `InstallStacksTable`
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/stacks`
- **Priority**: COMPLETE

#### 29. Install Stack Detail
- **Next.js**: `/app/[org-id]/installs/[install-id]/stacks/[stack-id]/page.tsx`
- **SPA**: `/pages/installs/InstallStackDetail.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Stack configuration
  - Terraform state
  - Recent operations
  - Drift detection
- **Components**: Stack config display, operations table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/stacks/{stackId}`
- **Priority**: LOW

#### 30. Install Runner
- **Next.js**: `/app/[org-id]/installs/[install-id]/runner/page.tsx`
- **SPA**: `/pages/installs/InstallRunner.tsx`
- **Status**: ✅ Migrated with functionality
- **Features**:
  - Install-specific runner info
  - Runner health
  - Recent jobs
  - Configuration
- **Components**: Runner health card, jobs table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/runner`
- **Priority**: COMPLETE

#### 31. Install Sandbox Overview
- **Next.js**: `/app/[org-id]/installs/[install-id]/sandbox/page.tsx`
- **SPA**: `/pages/installs/InstallSandbox.tsx`
- **Status**: ⚠️ Migrated but has mixed patterns (needs cleanup)
- **Features**:
  - Sandbox configuration
  - Recent runs
  - Drift detection banner
  - Values file management
- **Components**: `AppSandboxConfig`, `SandboxRunsTimeline`, `DriftedBanner`
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/sandbox`
  - `GET /api/ctl-api/v1/installs/{installId}/sandbox/runs`
- **Priority**: MEDIUM - Cleanup needed

#### 32. Install Sandbox Run Detail
- **Next.js**: `/app/[org-id]/installs/[install-id]/sandbox/[run-id]/page.tsx`
- **SPA**: `/pages/installs/InstallSandboxRun.tsx`
- **Status**: ⚠️ Basic implementation - needs enhancement
- **Current**: Shows run status and metadata
- **Missing**:
  - Terraform plan/apply output
  - Drift detection results
  - Resource changes breakdown
  - Logs viewer
  - Approval workflow integration
- **Components**: Run status display, plan/apply viewer, drift results
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/sandbox/runs/{runId}`
  - `GET /api/ctl-api/v1/installs/{installId}/sandbox/runs/{runId}/plan`
  - `GET /api/ctl-api/v1/installs/{installId}/sandbox/runs/{runId}/logs`
- **Priority**: HIGH - Enhancement needed

#### 33. Install Configuration / Inputs
- **Next.js**: `/app/[org-id]/installs/[install-id]/config/page.tsx`
- **SPA**: `/pages/installs/InstallConfig.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Install-specific config values
  - Input overrides
  - Edit configuration
  - Config history
- **Components**: `EditInputs`, config history table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/config`
  - `PUT /api/ctl-api/v1/installs/{installId}/config`
- **Priority**: HIGH

#### 34. Install State Viewer
- **Next.js**: `/app/[org-id]/installs/[install-id]/state/page.tsx`
- **SPA**: `/pages/installs/InstallState.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Terraform state viewer
  - Resource list
  - State history
  - Download state
- **Components**: `ViewState`, state diff viewer
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/state`
- **Priority**: MEDIUM

#### 35. Install Audit History
- **Next.js**: `/app/[org-id]/installs/[install-id]/audit/page.tsx`
- **SPA**: `/pages/installs/InstallAudit.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Audit events table
  - Event filtering
  - User attribution
  - Timestamp sorting
- **Components**: `AuditHistory`, events table
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/audit`
- **Priority**: MEDIUM

#### 36. Install Settings
- **Next.js**: `/app/[org-id]/installs/[install-id]/settings/page.tsx`
- **SPA**: `/pages/installs/InstallSettings.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Install name/description
  - Connection settings
  - Feature flags
  - Danger zone (delete install)
- **Components**: Settings forms, confirmation modals
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}`
  - `PUT /api/ctl-api/v1/installs/{installId}`
- **Priority**: MEDIUM

#### 37. Install Deploy History
- **Next.js**: Possibly in nested route
- **SPA**: `/pages/installs/InstallDeploys.tsx`
- **Status**: ❌ Needs implementation
- **Features**:
  - Full deploy history across all components
  - Deploy status filtering
  - Timeline view
- **Components**: Deploys table, timeline
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/deploys`
- **Priority**: LOW

#### 38. Install Approval Plans
- **Next.js**: Likely in workflows or separate route
- **SPA**: `/pages/installs/InstallApprovals.tsx`
- **Status**: ❌ Needs implementation (if feature exists)
- **Features**:
  - Pending approvals
  - Approval history
  - Plan review
  - Approve/reject actions
- **Components**: Approvals table, plan viewer
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/approvals`
- **Priority**: LOW

#### 39. Install Secrets Management
- **Next.js**: Possibly integrated in config
- **SPA**: `/pages/installs/InstallSecrets.tsx`
- **Status**: ❌ Needs implementation (if separate from config)
- **Features**:
  - Secrets list (masked)
  - Add/update secrets
  - Secret rotation
- **Components**: Secrets table, secret editor
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/secrets`
- **Priority**: LOW

#### 40. Install VCS Connections
- **Next.js**: `/app/[org-id]/installs/[install-id]/vcs/page.tsx` or in settings
- **SPA**: `/pages/installs/InstallVCS.tsx`
- **Status**: ❌ Needs verification if separate page exists
- **Features**:
  - VCS connection details
  - Repository information
  - Branch mapping
  - Sync status
- **Components**: VCS connection card, sync status
- **APIs**: 
  - `GET /api/ctl-api/v1/installs/{installId}/vcs`
- **Priority**: LOW

---

## Implementation Priority Matrix

### CRITICAL (Must implement immediately - user explicitly requested)
1. **Team Management Page** - User reported placeholder only
2. **Runner / Builds Page** - User reported placeholder only

### HIGH (Core functionality needed for daily operations)
3. Install Configuration / Inputs
4. Install Sandbox Run Detail (enhancement)
5. App Overview / Dashboard
6. App Components List
7. App Builds List
8. Apps List Page (full implementation)
9. Installs List Page (full implementation)

### MEDIUM (Important but not blocking)
10. Install Component Detail (enhancement)
11. Install Action Detail (enhancement)
12. Install Sandbox Overview (cleanup)
13. Install State Viewer
14. Install Audit History
15. Install Settings
16. App Settings
17. App Configuration
18. Organization Settings

### LOW (Nice to have, less frequently used)
19. App Build Detail
20. App Component Detail
21. App Installs
22. Install Policy Detail
23. Install Role Detail
24. Install Stack Detail
25. Install Deploy History
26. Install Approval Plans
27. Install Secrets Management
28. Install VCS Connections

---

## Detailed Implementation Guides

### CRITICAL Priority: Team Management Page

**File**: `/services/dashboard-ui/src/pages/org/TeamPage.tsx`

**Current State**: Placeholder with only heading

**Reference**: `/services/dashboard-ui/src/app/[org-id]/team/page.tsx`

**Required Features**:

1. **Team Members Table**:
   - Display all accounts with access to org
   - Show account email, name, role
   - Show last activity timestamp
   - Actions: Remove member, Change role

2. **Pending Invitations Section**:
   - Display pending invites
   - Show invite email, role, sent date
   - Actions: Resend invite, Cancel invite

3. **Invite New Member Flow**:
   - Button to open invite modal
   - Modal with:
     - Email input (validation)
     - Role selector (Admin, Installer, Runner)
     - Optional message
     - Send invite button

4. **Role Management**:
   - Display role descriptions
   - Change role dropdown per member
   - Confirmation for role changes

**Data Fetching**:
```typescript
const { data: members } = usePolling<TAccount[]>({
  path: `/api/ctl-api/v1/orgs/${orgId}/accounts`,
  pollInterval: 30000,
  shouldPoll: true,
})

const { data: invites } = usePolling<TOrgInvite[]>({
  path: `/api/ctl-api/v1/orgs/${orgId}/invites`,
  pollInterval: 30000,
  shouldPoll: true,
})
```

**Components to Build/Reuse**:
- `TeamMembersTable` component
- `InviteMemberModal` component
- `RoleSelector` component
- Confirmation modals for destructive actions

**API Endpoints**:
- `GET /api/ctl-api/v1/orgs/{orgId}/accounts` - List members
- `GET /api/ctl-api/v1/orgs/{orgId}/invites` - List pending invites
- `POST /api/ctl-api/v1/orgs/{orgId}/invites` - Send new invite
- `DELETE /api/ctl-api/v1/orgs/{orgId}/accounts/{accountId}` - Remove member
- `PUT /api/ctl-api/v1/orgs/{orgId}/accounts/{accountId}/role` - Change role
- `DELETE /api/ctl-api/v1/orgs/{orgId}/invites/{inviteId}` - Cancel invite
- `POST /api/ctl-api/v1/orgs/{orgId}/invites/{inviteId}/resend` - Resend invite

**Testing Checklist**:
- [ ] Page loads with team members table
- [ ] Pending invites section displays correctly
- [ ] Can open invite modal
- [ ] Can send invite with validation
- [ ] Can change member role
- [ ] Can remove member with confirmation
- [ ] Can cancel pending invite
- [ ] Can resend invite
- [ ] Polling updates data automatically
- [ ] Error states display correctly

---

### CRITICAL Priority: Runner / Builds Page

**File**: `/services/dashboard-ui/src/pages/org/OrgRunner.tsx`

**Current State**: Placeholder with only heading

**Reference**: `/services/dashboard-ui/src/app/[org-id]/runner/page.tsx`

**Required Features**:

1. **Runner Health Overview**:
   - Runner status (online/offline)
   - Health indicators
   - Last heartbeat timestamp
   - Runner version
   - Resource utilization

2. **Recent Jobs List**:
   - Jobs table with pagination
   - Job ID, type, status
   - Start/end timestamps
   - Duration
   - Associated install/app
   - Quick link to job details

3. **Runner Configuration**:
   - Runner settings display
   - Capacity limits
   - Enabled features
   - Connection details

4. **Performance Metrics** (optional):
   - Jobs per hour
   - Success rate
   - Average duration
   - Queue depth

**Data Fetching**:
```typescript
const { data: runner } = usePolling<TRunner>({
  path: `/api/ctl-api/v1/orgs/${orgId}/runner`,
  pollInterval: 20000,
  shouldPoll: true,
})

const { data: jobs } = usePolling<TRunnerJob[]>({
  path: `/api/ctl-api/v1/orgs/${orgId}/runner/jobs?limit=20`,
  pollInterval: 20000,
  shouldPoll: true,
})

const { data: health } = usePolling<TRunnerHealth>({
  path: `/api/ctl-api/v1/orgs/${orgId}/runner/health`,
  pollInterval: 10000,
  shouldPoll: true,
})
```

**Components to Build/Reuse**:
- `RunnerHealthCard` component (may already exist)
- `RunnerJobsTable` component
- `RunnerConfigPanel` component
- Health status indicators
- Duration/timestamp formatters

**API Endpoints**:
- `GET /api/ctl-api/v1/orgs/{orgId}/runner` - Get runner details
- `GET /api/ctl-api/v1/orgs/{orgId}/runner/health` - Health check
- `GET /api/ctl-api/v1/orgs/{orgId}/runner/jobs` - List jobs
- `GET /api/ctl-api/v1/orgs/{orgId}/runner/jobs/{jobId}` - Job detail

**Testing Checklist**:
- [ ] Page loads with runner health status
- [ ] Recent jobs table displays correctly
- [ ] Job status indicators work
- [ ] Can navigate to job details
- [ ] Health indicators update via polling
- [ ] Timestamps format correctly
- [ ] Duration calculations correct
- [ ] Pagination works for jobs
- [ ] Loading states display correctly
- [ ] Error states display correctly

---

### HIGH Priority: Install Configuration / Inputs

**File**: `/services/dashboard-ui/src/pages/installs/InstallConfig.tsx`

**Status**: ❌ Needs creation

**Reference**: Component exists: `/services/dashboard-ui/src/components/installs/management/EditInputs.tsx`

**Required Features**:

1. **Configuration Display**:
   - List all config inputs
   - Show current values (masked for secrets)
   - Show default values
   - Show input type and validation

2. **Edit Mode**:
   - Toggle edit mode
   - Input editors by type (text, number, boolean, select)
   - Validation feedback
   - Save/cancel buttons

3. **Config History** (optional):
   - Show previous config versions
   - Timestamp and user who changed
   - Diff viewer

**Implementation**:
```typescript
export default function InstallConfig() {
  const { orgId, installId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: config } = usePolling<TInstallConfig>({
    path: `/api/ctl-api/v1/installs/${installId}/config`,
    pollInterval: 30000,
    shouldPoll: true,
  })

  return (
    <PageSection isScrollable>
      <Breadcrumbs breadcrumbs={[...]} />
      <HeadingGroup>
        <Text variant="h3">Configuration</Text>
      </HeadingGroup>
      
      <EditInputs 
        installId={installId}
        initialConfig={config}
      />
      
      <BackToTop />
    </PageSection>
  )
}
```

---

### HIGH Priority: Install Sandbox Run Detail (Enhancement)

**File**: `/services/dashboard-ui/src/pages/installs/InstallSandboxRun.tsx`

**Current State**: Shows basic metadata only

**Missing Features**:

1. **Terraform Plan Output**:
   - Fetch and display plan from API
   - Resource additions/changes/deletions
   - Plan diff viewer
   - Color coding for changes

2. **Terraform Apply Output**:
   - Apply results
   - Resource creation status
   - Error messages if failed

3. **Drift Detection**:
   - Drift results if available
   - Resources with drift
   - Drift details

4. **Logs Integration**:
   - Link to or embed log viewer
   - Filter logs for this run

5. **Approval Workflow**:
   - Show if approval required
   - Approve/reject buttons
   - Approval history

**Enhanced Implementation**:
```typescript
export default function InstallSandboxRun() {
  const { orgId, installId, runId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: sandboxRun } = usePolling<TSandboxRun>({
    path: `/api/ctl-api/v1/installs/${installId}/sandbox/runs/${runId}`,
    pollInterval: 20000,
    shouldPoll: true,
  })

  const { data: plan } = useQuery<TTerraformPlan>({
    path: `/api/ctl-api/v1/installs/${installId}/sandbox/runs/${runId}/plan`,
  })

  const { data: drift } = useQuery<TDriftResult>({
    path: `/api/ctl-api/v1/installs/${installId}/sandbox/runs/${runId}/drift`,
  })

  return (
    <PageSection isScrollable>
      <Breadcrumbs breadcrumbs={[...]} />
      <HeadingGroup>
        <div className="flex items-center gap-3">
          <Text variant="h3">Sandbox Run</Text>
          <Status status={sandboxRun?.status_v2?.status} variant="badge" />
        </div>
        <ID>{runId}</ID>
      </HeadingGroup>

      {/* Metadata Section */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Status, Created, Updated, Duration */}
      </div>

      {/* Plan Section */}
      {plan && (
        <div className="mt-6">
          <Text variant="base" weight="strong" className="mb-4">
            Terraform Plan
          </Text>
          <TerraformPlanViewer plan={plan} />
        </div>
      )}

      {/* Drift Section */}
      {drift && (
        <div className="mt-6">
          <Text variant="base" weight="strong" className="mb-4">
            Drift Detection
          </Text>
          <DriftResultsViewer drift={drift} />
        </div>
      )}

      {/* Logs Section */}
      <div className="mt-6">
        <Text variant="base" weight="strong" className="mb-4">
          Run Logs
        </Text>
        <Button onClick={() => navigate(`logs`)}>View Logs</Button>
      </div>

      <BackToTop />
    </PageSection>
  )
}
```

**New Components Needed**:
- `TerraformPlanViewer` - Display plan with resource changes
- `DriftResultsViewer` - Display drift detection results

**Additional API Endpoints**:
- `GET /api/ctl-api/v1/installs/{installId}/sandbox/runs/{runId}/plan`
- `GET /api/ctl-api/v1/installs/{installId}/sandbox/runs/{runId}/drift`
- `POST /api/ctl-api/v1/installs/{installId}/sandbox/runs/{runId}/approve`

---

## Testing Strategy

### Automated Testing
1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test page-level data flow
3. **E2E Tests**: Test complete user journeys

### Manual Testing with Chrome MCP
For each implemented page:
1. Navigate to page at `localhost:4000`
2. Verify page loads without console errors
3. Verify data fetches correctly
4. Verify polling updates data
5. Test all interactive elements (buttons, forms, etc.)
6. Test error states (network failures, 404s, etc.)
7. Test loading states
8. Test on different screen sizes (responsive design)

### Testing Restart Mechanism
```bash
# After code changes
touch ~/.nuonctl-restart-dashboard-ui

# Wait 5-10 seconds for service restart
# Then test in browser
```

### Chrome MCP Testing Commands
```typescript
// Navigate to page
navigate_page({ url: 'http://localhost:4000/{orgId}/team' })

// Take snapshot to verify UI
take_snapshot()

// Check for console errors
list_console_messages({ types: ['error'] })

// Click elements to test interactions
click({ uid: 'button-uid-from-snapshot' })

// Verify data loads
evaluate_script({ 
  function: '() => document.body.innerText.includes("Expected Text")'
})
```

---

## Implementation Order Recommendation

**Phase 1: CRITICAL (Week 1)**
1. Team Management Page - Full implementation
2. Runner / Builds Page - Full implementation
3. Test both extensively with Chrome MCP

**Phase 2: HIGH Priority (Week 2-3)**
4. Install Configuration / Inputs - New page
5. Install Sandbox Run Detail - Enhancement
6. Install Component Detail - Enhancement
7. Install Action Detail - Enhancement
8. Apps List Page - Full implementation
9. Installs List Page - Full implementation

**Phase 3: App Pages (Week 4-5)**
10. App Overview / Dashboard
11. App Components List
12. App Builds List
13. App Settings
14. App Configuration

**Phase 4: MEDIUM Priority (Week 6-7)**
15. Install Sandbox Overview - Cleanup
16. Install State Viewer
17. Install Audit History
18. Install Settings
19. Organization Settings

**Phase 5: LOW Priority (Week 8+)**
20. All remaining detail pages
21. Optional/advanced features
22. Performance optimizations
23. Accessibility improvements

---

## Common Patterns & Best Practices

### 1. Data Fetching Pattern
```typescript
// Use usePolling for real-time data
const { data, isLoading, error } = usePolling<TDataType>({
  path: `/api/ctl-api/v1/...`,
  pollInterval: 20000, // 20s for frequently changing data
  shouldPoll: true,
})

// Use useQuery for one-time data
const { data } = useQuery<TDataType>({
  path: `/api/ctl-api/v1/...`,
})
```

### 2. Loading States
```typescript
if (isLoading) {
  return (
    <PageSection isScrollable>
      <Loading variant="stack" loadingText="Loading..." />
    </PageSection>
  )
}
```

### 3. Error States
```typescript
if (error) {
  return (
    <PageSection isScrollable>
      <Text theme="error">Failed to load data: {error.message}</Text>
    </PageSection>
  )
}
```

### 4. Empty States
```typescript
if (!data || data.length === 0) {
  return (
    <PageSection isScrollable>
      <Text theme="neutral">No items found.</Text>
    </PageSection>
  )
}
```

### 5. Breadcrumbs Pattern
```typescript
<Breadcrumbs
  breadcrumbs={[
    { path: `/${orgId}`, text: org?.name || '' },
    { path: `/${orgId}/section`, text: 'Section' },
    { path: `/${orgId}/section/detail`, text: 'Detail' },
  ]}
/>
```

### 6. Provider Access
```typescript
const { org } = useOrg() // Current org context
const { install } = useInstall() // Current install context
const { app } = useApp() // Current app context
```

### 7. Navigation
```typescript
import { useNavigate } from 'react-router-dom'

const navigate = useNavigate()
navigate(`/${orgId}/path`)
```

### 8. Feature Flags
```typescript
const { org } = useOrg()
const hasFeature = org?.feature_flags?.includes('feature-name')

if (!hasFeature) {
  return <Text>Feature not enabled</Text>
}
```

---

## API Endpoint Reference

All endpoints are proxied through `/api/ctl-api/v1/` prefix.

### Organization Level
- `GET /orgs/{orgId}` - Org details
- `GET /orgs/{orgId}/apps` - List apps
- `GET /orgs/{orgId}/installs` - List installs
- `GET /orgs/{orgId}/accounts` - List team members
- `GET /orgs/{orgId}/invites` - List pending invites
- `POST /orgs/{orgId}/invites` - Send invite
- `GET /orgs/{orgId}/runner` - Runner details
- `GET /orgs/{orgId}/runner/jobs` - Runner jobs

### App Level
- `GET /apps/{appId}` - App details
- `GET /apps/{appId}/components` - List components
- `GET /apps/{appId}/builds` - List builds
- `POST /apps/{appId}/builds` - Trigger build
- `GET /apps/{appId}/config` - App config

### Install Level
- `GET /installs/{installId}` - Install details
- `GET /installs/{installId}/components` - List components
- `GET /installs/{installId}/components/{componentId}` - Component detail
- `GET /installs/{installId}/action-workflows` - List actions
- `GET /installs/{installId}/action-workflows/{actionId}` - Action detail
- `GET /installs/{installId}/workflows` - List workflows
- `GET /installs/{installId}/workflows/{workflowId}` - Workflow detail
- `GET /installs/{installId}/sandbox` - Sandbox config
- `GET /installs/{installId}/sandbox/runs` - Sandbox runs
- `GET /installs/{installId}/sandbox/runs/{runId}` - Run detail
- `GET /installs/{installId}/config` - Install config
- `PUT /installs/{installId}/config` - Update config

---

## Migration Tracking

Use this checklist to track progress:

### CRITICAL Priority
- [ ] Team Management Page - Full implementation
- [ ] Runner / Builds Page - Full implementation

### HIGH Priority
- [ ] Install Configuration / Inputs
- [ ] Install Sandbox Run Detail - Enhancement
- [ ] Install Component Detail - Enhancement
- [ ] Install Action Detail - Enhancement
- [ ] Apps List Page - Full implementation
- [ ] Installs List Page - Full implementation
- [ ] App Overview / Dashboard
- [ ] App Components List
- [ ] App Builds List

### MEDIUM Priority
- [ ] Install Sandbox Overview - Cleanup
- [ ] Install State Viewer
- [ ] Install Audit History
- [ ] Install Settings
- [ ] Organization Settings
- [ ] App Settings
- [ ] App Configuration

### LOW Priority
- [ ] App Build Detail
- [ ] App Component Detail
- [ ] App Installs
- [ ] Install Policy Detail
- [ ] Install Role Detail
- [ ] Install Stack Detail
- [ ] Install Deploy History
- [ ] Install Approval Plans (if exists)
- [ ] Install Secrets Management (if separate)
- [ ] Install VCS Connections (if separate)

---

## Notes and Warnings

1. **DO NOT modify ctl-api backend** - All endpoints already exist
2. **DO NOT modify proxy or auth** - Cookie handling already works
3. **Follow existing patterns** - Reference working pages like InstallOverview
4. **Test extensively** - Use Chrome MCP before asking user to test
5. **Placeholder pages are NOT sufficient** - User expects full Next.js feature parity
6. **Polling intervals**: 10-20s for high-frequency data, 30s for lower frequency
7. **Feature flags**: Check org.feature_flags before rendering certain features
8. **Error handling**: Always display errors gracefully, never crash
9. **Loading states**: Always show loading indicator during data fetch
10. **Responsive design**: Test on different screen sizes

---

## Questions to Resolve

1. **Install Approval Plans**: Confirm if this is a separate page or integrated into sandbox runs
2. **Install Secrets**: Confirm if secrets are separate from config or integrated
3. **Install VCS Connections**: Confirm if this is a separate page or in settings
4. **API Endpoints**: Verify all endpoint paths match actual ctl-api routes
5. **Feature Flags**: Confirm which features are gated by flags
6. **Permissions**: Confirm if any pages require role-based access control

---

## Success Criteria

A page is considered "complete" when:

1. ✅ Page loads without errors
2. ✅ Data fetches correctly from API
3. ✅ Polling updates data automatically
4. ✅ All interactive elements work (buttons, forms, etc.)
5. ✅ Loading states display correctly
6. ✅ Error states display gracefully
7. ✅ Empty states display appropriately
8. ✅ Breadcrumbs navigate correctly
9. ✅ Feature parity with Next.js version
10. ✅ Responsive design works on mobile/tablet/desktop
11. ✅ No console errors in browser
12. ✅ Manual testing with Chrome MCP passes
13. ✅ User testing passes

---

## Conclusion

This plan covers all 40+ pages identified in the Next.js app directory. The priority matrix ensures critical user-reported issues are addressed first, followed by high-value features, then lower-priority detail pages.

The key insight from user feedback is that **placeholder implementations are not sufficient** - each page must have full functional parity with the Next.js version, including all data fetching, interactive elements, and sub-features.

By following this plan systematically and testing thoroughly with Chrome MCP before user testing, we can ensure a smooth migration with minimal iteration cycles.

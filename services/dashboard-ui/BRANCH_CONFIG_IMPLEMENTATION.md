# Branch Configuration Implementation Summary

## Overview

This document summarizes the complete frontend implementation for the branch configuration management system, including branch configs, install groups, and workflow runs.

## Implementation Date

2026-01-21

## Components Implemented

### 1. API Client Functions (`/src/lib/ctl-api/apps/branches/`)

**New API client functions:**
- `get-branch-configs.ts` - Fetch all configurations for a branch
- `get-branch-config.ts` - Fetch a specific configuration
- `create-branch-config.ts` - Create a new configuration version
- `update-branch.ts` - Update branch name
- `get-branch-workflow-runs.ts` - Fetch workflow runs for a branch

**Types added to `/src/types/ctl-api.types.ts`:**
- `TAppBranchConfig` - Branch configuration structure
- `TAppBranchInstallGroup` - Install group configuration
- Exported from OpenAPI schema

### 2. EditBranchNamePanel Component

**File:** `/src/app/[org-id]/apps/[app-id]/branches/[branch-id]/edit-branch-name-panel.tsx`

**Features:**
- Modal panel for editing branch name
- Size: "half" (simple edit form)
- Info banner explaining limitations
- Real-time validation
- Error handling with user feedback
- Auto-refresh on successful save

**Usage:**
```typescript
<EditBranchNamePanel
  branch={branch}
  orgId={orgId}
  appId={appId}
  isVisible={showEditName}
  onClose={() => setShowEditName(false)}
/>
```

### 3. NewBranchConfigPanel Component

**File:** `/src/app/[org-id]/apps/[app-id]/branches/[branch-id]/new-branch-config-panel.tsx`

**Features:**
- Modal panel for creating new configuration versions
- Size: "3/4" (complex form with multiple sections)
- VCS Configuration section:
  - VCS connection selector
  - Repository dropdown (fetched from GitHub)
  - Branch dropdown (fetched from GitHub)
  - Directory path input
  - Path filter (regex) input
- Install Groups section:
  - Dynamic group management (add/remove)
  - Multi-select install picker
  - Group ordering
  - Max parallel setting
  - Requires approval checkbox
  - Rollback on failure checkbox
- Pre-populated with current config values
- Shows next version number in header
- Comprehensive validation
- Error handling with user feedback

**Usage:**
```typescript
<NewBranchConfigPanel
  branch={branch}
  currentConfig={currentConfig}
  vcsConnections={vcsConnections}
  orgId={orgId}
  appId={appId}
  isVisible={showNewConfig}
  onClose={() => setShowNewConfig(false)}
/>
```

### 4. BranchConfigHistoryTable Component

**File:** `/src/app/[org-id]/apps/[app-id]/branches/[branch-id]/branch-config-history-table.tsx`

**Features:**
- Table displaying all configuration versions
- Columns:
  - Version (with CURRENT badge for active config)
  - Repository
  - Branch
  - Install Groups count
  - Created timestamp (relative)
- Client-side sorting and search
- Loading skeleton
- Error handling
- Empty state
- Auto-fetches and sorts by version (newest first)

**Usage:**
```typescript
<BranchConfigHistoryTable
  appId={appId}
  branchId={branchId}
  orgId={orgId}
  currentConfigId={currentConfig?.id}
/>
```

### 5. BranchWorkflowRunsTable Component

**File:** `/src/app/[org-id]/apps/[app-id]/branches/[branch-id]/branch-workflow-runs-table.tsx`

**Features:**
- Table displaying workflow runs for the branch
- Columns:
  - Run ID (short code with link)
  - Status (with Status component)
  - Config version
  - Install name
  - Started timestamp (relative)
  - Duration
  - View action link
- Client-side sorting and search
- Loading skeleton
- Error handling
- Empty state
- Links to workflow detail pages

**Usage:**
```typescript
<BranchWorkflowRunsTable
  appId={appId}
  branchId={branchId}
  orgId={orgId}
/>
```

### 6. BranchDetailActions Component

**File:** `/src/app/[org-id]/apps/[app-id]/branches/[branch-id]/branch-detail-actions.tsx`

**Features:**
- Action buttons for branch detail page
- "Edit Name" button → Opens EditBranchNamePanel
- "New Config Version" button → Opens NewBranchConfigPanel
- "Trigger Run" button (currently disabled, ready for future implementation)
- Manages modal state

**Usage:**
```typescript
<BranchDetailActions
  branch={branch}
  currentConfig={currentConfig}
  vcsConnections={vcsConnections}
  appId={appId}
  orgId={orgId}
/>
```

### 7. Branch Detail Page (New Implementation)

**File:** `/src/app/[org-id]/apps/[app-id]/branches/[branch-id]/page-new.tsx`

**Features:**
- Complete branch detail view with three main sections
- Page header with branch name, ID, and created timestamp
- Action buttons (Edit Name, New Config, Trigger Run)
- Current Configuration section:
  - Version badge
  - Creation timestamp
  - Install groups count
  - VCS configuration details (repo, branch, directory, path filter)
  - Install groups details (name, install count, max parallel, flags)
  - Empty state when no config exists
- Configuration History section:
  - Uses BranchConfigHistoryTable
  - Wrapped in ErrorBoundary + Suspense
- Workflow Runs section:
  - Uses BranchWorkflowRunsTable
  - Wrapped in ErrorBoundary + Suspense
- Breadcrumb navigation
- Server-side rendering with async data fetching

**Data Flow:**
```
Server:
- getAppBranch() → branch data with configs
- getApp() → app details
- getOrg() → org details
- getVCSConnections() → available VCS connections
- Sort configs to find current (highest version number)

Client:
- EditBranchNamePanel → Modal for editing name
- NewBranchConfigPanel → Modal for creating new config
- Tables fetch their own data client-side
```

## API Endpoints Used

All endpoints are in the ctl-api service:

**Branch Management:**
- `GET /v1/apps/{app_id}/branches/{branch_id}` - Get branch details
- `PATCH /v1/apps/{app_id}/branches/{branch_id}` - Update branch name
- `GET /v1/apps/{app_id}/branches/{branch_id}/configs` - List configs
- `POST /v1/apps/{app_id}/branches/{branch_id}/configs` - Create config
- `GET /v1/apps/{app_id}/branches/{branch_id}/workflow-runs` - List workflow runs

**Supporting Endpoints:**
- `GET /v1/vcs/connections/{connection_id}/repos` - List GitHub repos
- `GET /v1/vcs/connections/{connection_id}/repos/{owner}/{repo}/branches` - List branches
- `GET /v1/apps/{app_id}/installs` - List installs (for install groups picker)

## Data Structures

### AppBranchConfig
```typescript
{
  id: string
  app_branch_id: string
  config_number: number  // Auto-incremented version
  created_at: string
  created_by_id: string
  org_id: string
  connected_github_vcs_config?: {
    vcs_connection_id: string
    repo: string
    branch: string
    directory: string
    path_filter?: string
  }
  install_groups?: AppBranchInstallGroup[]
  workflows?: Workflow[]
}
```

### AppBranchInstallGroup
```typescript
{
  id: string
  app_branch_config_id: string
  name: string
  install_ids: string[]
  order: number
  max_parallel: number
  requires_approval: boolean
  rollback_on_failure: boolean
  created_at: string
  created_by_id: string
  org_id: string
}
```

## User Workflows

### Creating a New Configuration Version

1. User clicks "New Config Version" button
2. NewBranchConfigPanel opens with current config pre-populated
3. User can modify:
   - VCS connection
   - Repository
   - Branch
   - Directory
   - Path filter
   - Install groups (add/remove/edit)
4. User clicks "Create Configuration v{X}"
5. API creates new config with incremented version number
6. Page refreshes showing new config as current
7. Old config remains in history table

### Editing Branch Name

1. User clicks "Edit Name" button
2. EditBranchNamePanel opens with current name
3. User modifies name
4. User clicks "Save Changes"
5. API updates branch name
6. Page refreshes with new name
7. Panel closes

### Viewing Configuration History

1. User scrolls to "Configuration History" section
2. Table loads all configs for the branch
3. Current config is badged with "CURRENT"
4. Configs sorted by version (newest first)
5. User can see repo, branch, group count, and created time for each version

### Viewing Workflow Runs

1. User scrolls to "Workflow Runs" section
2. Table loads all workflow runs for the branch
3. Runs sorted by created timestamp (newest first)
4. User can see run ID, status, config version, install name, started time, and duration
5. User clicks "View" to go to workflow detail page

## File Structure

```
/services/dashboard-ui/src/
├── lib/ctl-api/apps/branches/
│   ├── get-branch-configs.ts
│   ├── get-branch-config.ts
│   ├── create-branch-config.ts
│   ├── update-branch.ts
│   ├── get-branch-workflow-runs.ts
│   └── index.ts (exports all)
├── types/
│   └── ctl-api.types.ts (TAppBranchConfig, TAppBranchInstallGroup)
└── app/[org-id]/apps/[app-id]/branches/[branch-id]/
    ├── page-new.tsx (NEW branch detail page)
    ├── edit-branch-name-panel.tsx
    ├── new-branch-config-panel.tsx
    ├── branch-detail-actions.tsx
    ├── branch-config-history-table.tsx
    └── branch-workflow-runs-table.tsx
```

## Integration Points

### Existing Components Used
- `Modal` (`/src/components/surfaces/Modal.tsx`) - For panels
- `Table` (`/src/components/common/Table.tsx`) - For tables
- `Badge` (`/src/components/common/Badge.tsx`) - For status badges
- `Banner` (`/src/components/common/Banner.tsx`) - For messages
- `Card` (`/src/components/common/Card.tsx`) - For sections
- `Status` (`/src/components/common/Status.tsx`) - For workflow status
- `Time` (`/src/components/common/Time.tsx`) - For timestamps
- `Duration` (`/src/components/common/Duration.tsx`) - For durations
- Various form components (Input, Button, etc.)

### Hooks Used
- `useRouter` - Navigation and refresh
- `useState` - Local state management
- `useEffect` - Data fetching and side effects

### External APIs Used
- VCS Connections API (for repo/branch fetching)
- Apps API (for install fetching)
- Branches API (all CRUD operations)

## Testing Considerations

### Manual Testing Checklist

**Branch Name Edit:**
- [ ] Opens modal on "Edit Name" click
- [ ] Pre-populates current name
- [ ] Validates empty name
- [ ] Shows error on API failure
- [ ] Refreshes page on success
- [ ] Closes modal on success
- [ ] Closes modal on cancel

**New Config Creation:**
- [ ] Opens modal on "New Config Version" click
- [ ] Pre-populates current config values
- [ ] Fetches repositories when VCS connection changes
- [ ] Fetches branches when repository changes
- [ ] Allows adding/removing install groups
- [ ] Validates required fields
- [ ] Shows error on API failure
- [ ] Refreshes page on success
- [ ] Closes modal on success
- [ ] Displays correct version number

**Config History Table:**
- [ ] Loads all configurations
- [ ] Badges current config
- [ ] Sorts by version descending
- [ ] Shows loading skeleton
- [ ] Shows error banner on failure
- [ ] Shows empty state when no configs
- [ ] Search filters correctly
- [ ] Sorting works correctly

**Workflow Runs Table:**
- [ ] Loads all workflow runs
- [ ] Displays status correctly
- [ ] Links to workflow detail pages
- [ ] Shows loading skeleton
- [ ] Shows error banner on failure
- [ ] Shows empty state when no runs
- [ ] Search filters correctly
- [ ] Sorting works correctly

**Current Config Display:**
- [ ] Shows correct version number
- [ ] Displays VCS configuration
- [ ] Lists install groups with details
- [ ] Shows flags (approval, rollback)
- [ ] Shows empty state when no config

## Future Enhancements

### Planned Features
1. **Trigger Run Button** - Implement workflow run triggering from branch detail page
2. **Config Diff View** - Compare two config versions side-by-side
3. **Config Rollback** - Ability to rollback to a previous config version
4. **Workflow Run Filtering** - Filter runs by status, config version, install
5. **Bulk Actions** - Apply config to multiple branches at once
6. **Config Templates** - Save common configs as templates
7. **Approval Workflow** - Manual approval process for config changes
8. **Audit Log** - Track who made changes and when
9. **Config Validation** - Pre-validate configs before creation
10. **Dry Run** - Test a config without actually deploying

### Technical Improvements
1. **Optimistic Updates** - Update UI before API response
2. **Better Error Messages** - More specific error messages
3. **Loading States** - More granular loading indicators
4. **Caching** - Cache config history and workflow runs
5. **Pagination** - Paginate workflow runs table
6. **Real-time Updates** - Poll for new workflow runs
7. **Keyboard Shortcuts** - Add keyboard navigation
8. **Export** - Export config history as JSON/YAML
9. **Import** - Import configs from file
10. **Validation** - Validate regex path filters

## Known Limitations

1. **No Config Deletion** - Configs are append-only (by design)
2. **No Inline Editing** - Must create new version to change config
3. **No Diff View** - Cannot compare configs visually
4. **No Rollback** - Cannot easily revert to previous config
5. **Manual Refresh** - Tables don't auto-refresh
6. **No Pagination** - Large history/runs lists load all at once
7. **Limited Validation** - Path filter regex not validated before save
8. **No Dry Run** - Cannot test config before applying
9. **Trigger Run Disabled** - Workflow triggering not yet implemented
10. **No Config Search** - Cannot search within config contents

## Dependencies

### Backend Requirements
- ctl-api endpoints implemented and working
- VCS connections configured in org
- Installs created for the app
- Auth0 authentication working

### Frontend Requirements
- Next.js 15+ with App Router
- TypeScript configured
- Tailwind CSS configured
- Modal/Table components available
- VCS connections API integrated

## Deployment Notes

1. **No Migration Required** - All data structures exist in backend
2. **Feature Flag** - Consider gating behind org feature flag
3. **Progressive Rollout** - Can roll out to subset of orgs first
4. **Backward Compatible** - Existing branch list still works
5. **No Breaking Changes** - All new components, no modifications to existing

## Success Metrics

### User Experience
- Time to create new config version
- Number of config versions per branch
- Error rate on config creation
- Modal abandonment rate
- Search/filter usage in tables

### Technical
- API response times
- Error rates by endpoint
- Client-side performance
- Modal render time
- Table load time

## Documentation Links

- [Backend API Spec](http://localhost:8081/oapi/v3)
- [Branch Configs PR](#) - TODO: Add PR link
- [Design Mockups](#) - TODO: Add design link
- [Original RFC](#) - TODO: Add RFC link

## Support

For questions or issues:
- Frontend: @dashboard-ui team
- Backend: @ctl-api team
- Product: @product team

## Changelog

### 2026-01-21 - Initial Implementation
- Created all API client functions
- Implemented EditBranchNamePanel
- Implemented NewBranchConfigPanel
- Implemented BranchConfigHistoryTable
- Implemented BranchWorkflowRunsTable
- Implemented BranchDetailActions
- Implemented complete Branch Detail Page
- Added types to ctl-api.types.ts
- Documented all components and workflows

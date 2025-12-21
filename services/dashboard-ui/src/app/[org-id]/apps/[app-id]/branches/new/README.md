# App Branch Onboarding Wizard

A complete multi-step wizard for configuring app branch deployment strategies with stub data only.

## Overview

This wizard guides users through creating a new app branch configuration with:
- Basic information (name, description)
- VCS integration (optional)
- Install grouping with deployment order
- Review and confirmation

All data is stored in localStorage for testing without API integration.

## Files Structure

```
branches/new/
├── page.tsx                      # Main wizard orchestration
├── types.ts                      # TypeScript interfaces
├── mock-data.ts                  # Mock data and localStorage utilities
├── step-basic-info.tsx           # Step 1: Name and description
├── step-vcs-config.tsx           # Step 2: VCS configuration
├── step-install-groups.tsx       # Step 3: Install grouping
├── step-review.tsx               # Step 4: Review and create
├── path-filter-validator.tsx     # Path filter testing component
├── install-card.tsx              # Reusable install card component
└── README.md                     # This file
```

## Features

### Step 1: Basic Info
- Branch name (required)
- Description (optional)

### Step 2: VCS Configuration
- Toggle for manual-only mode (no VCS)
- VCS connection selector
- Repository selector
- Git branch selector
- Directory input
- **Path Filter Validator** with:
  - Regex pattern input
  - Real-time validation
  - Pattern tester with sample paths
  - Common pattern examples dropdown
  - Vercel-style monorepo support

### Step 3: Install Groups
- **Template Selection**:
  - Production → Staging
  - Canary Deployment
  - Regional Rollout
  - Custom (blank)
  
- **Interactive Grouping**:
  - Drag installs into groups (via arrow buttons)
  - Reorder groups with up/down controls
  - Search/filter available installs
  
- **Group Settings**:
  - Custom group names
  - Require manual approval
  - Wait for health checks
  - Auto rollback on failure
  - Max parallel installs

### Step 4: Review & Create
- Summary of all configuration
- "What happens next" info panel
- Group settings badges
- Create button (saves to localStorage)

## Mock Data

### Available Installs (10 total)
- production-us-east, production-us-west
- staging-us-east, staging-us-west
- dev-environment, qa-environment
- demo-environment, sandbox-1, sandbox-2
- test-env (inactive)

### VCS Connections
- github-org
- github-personal

### Repositories
- nuonco/app-backend (private)
- nuonco/app-frontend (public)
- nuonco/infrastructure (private)

## localStorage Schema

```typescript
// Key pattern
key: `app-branch-config-${branchId}`

// Value structure
{
  name: string
  description?: string
  vcsEnabled: boolean
  vcsConnectionId?: string
  repository?: string
  gitBranch?: string
  directory?: string
  pathFilter?: string
  installGroups: [
    {
      id: string
      name: string
      installIds: string[]
      order: number
      requiresApproval: boolean
      rollbackOnFailure: boolean
      maxParallel: number
    }
  ]
}
```

## Usage

1. Navigate to `/[org-id]/apps/[app-id]/branches/new`
2. Complete the 4-step wizard
3. Click "Create Branch" on final step
4. Redirects to canvas view with mock data loaded from localStorage

## Testing

To test the wizard:

1. Start the dashboard: `npm run dev`
2. Navigate to any app's branches page
3. Click "Create New Branch"
4. Follow the wizard steps
5. Create the branch
6. Check browser localStorage for saved configuration
7. Navigate to the canvas view to see the configuration visualized

## Path Filter Examples

The path filter validator includes these common patterns:

- **All files**: `` (empty)
- **Source code only**: `^(src/|lib/)`
- **Ignore docs**: `^(?!docs/)`
- **Specific file types**: `\.(ts|tsx|js|jsx)$`
- **Monorepo package**: `^packages/frontend/`
- **Config files**: `^(config/|.env|\.yaml$|\.json$)`

## Group Templates

### Production → Staging
- Staging (no approval, health checks, rollback)
- Production (requires approval, health checks, rollback)

### Canary Deployment
- Canary (1 install, no approval)
- Wave 1 (3 installs, requires approval)
- Wave 2 (10 installs, no approval)

### Regional Rollout
- US East (5 parallel)
- US West (5 parallel)
- Europe (5 parallel)

### Custom
- Empty template for manual configuration

## Future Enhancements

When connecting to the real API:

1. Replace mock data functions with API calls
2. Remove localStorage persistence
3. Add proper validation errors from backend
4. Implement actual branch creation endpoint
5. Add edit mode for existing branches
6. Integrate with canvas view for real workflow runs

## Component Dependencies

- `@/components/common/Button`
- `@/components/common/Text`
- `@/components/common/Icon`
- `@/components/common/Badge`
- `@/components/common/Card`
- `@/components/common/Banner`
- `@/components/old/Input`
- `@/components/old/Dropdown`
- `@/components/layout/PageSection`
- `@/components/navigation/Breadcrumb`

All components are existing in the codebase.
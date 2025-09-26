# Nuon Monorepo Overview

This is a comprehensive monorepo for Nuon, a BYOC (Bring Your Own Cloud) platform that helps software vendors deploy applications to their customers' cloud accounts.

## Repository Structure

This monorepo is written primarily in **Go** (main module: `github.com/powertoolsdev/mono`) with several TypeScript/JavaScript projects for UI components.

### Core Directories

#### `/bins/` - Command Line Tools & Executables
Binaries compiled and run as executables (not deployed as Kubernetes services):
- **`cli/`** - Public-facing CLI tool for developers (`nuon` command)
- **`nuonctl/`** - Internal CLI with 80+ operational scripts and dev tooling
- **`runner/`** - Deployment execution binary (runs in customer K8s + VMs)

#### `/services/` - Web Services & Applications
- **`ctl-api/`** - Control API service (Go)
- **`dashboard-ui/`** - Main dashboard UI (Next.js/React)
- **`website-v2/`** - Marketing website (Astro)
- **`website/`** - Legacy website (Astro)
- **`wiki/`** - Internal documentation site (Astro + Starlight)
- **`e2e/`** - End-to-end testing service
- **`orgs-api/`** - Organizations API service (DEPRECATED)
- **`workers-*`** - Various worker services for background tasks
  - **`workers-executors/`** - (DEPRECATED) Previously core execution engine
  - **`workers-canary/`** - Canary testing service
  - **`workers-infra-tests/`** - Infrastructure testing service

#### `/pkg/` - Shared Go Libraries
Core shared packages used across Go services:
- `api/` - API client libraries
- `config/` - Configuration management
- `helm/` - Helm chart operations
- `kube/` - Kubernetes utilities
- `terraform/` - Terraform operations
- `metrics/` - Metrics and telemetry
- `temporal/` - Temporal workflow engine integration

#### `/infra/` - Infrastructure as Code
Terraform modules and configurations for:
- AWS infrastructure
- Azure support
- Kubernetes (EKS) clusters
- Monitoring (DataDog)
- DNS management
- Developer environments

#### `/charts/` - Helm Charts
Kubernetes Helm charts for various components:
- `common/` - Shared chart templates
- `temporal/` - Temporal workflow engine
- Application-specific charts

#### `/docs/` - Documentation
- API reference documentation
- User guides and tutorials
- Platform support documentation
- Production readiness guides

#### `/seed/` - Example Applications
Template applications and configurations for different deployment scenarios:
- EKS examples
- Azure AKS examples
- Component-based applications
- Various infrastructure patterns

#### `/exp/` - Experimental Features
Development and testing area for new features and proof-of-concepts

#### `/graveyard/` - Deprecated Code
Archive of deprecated code and components

#### `/images/` - Container Images
Dockerfiles and build configurations for various container images

#### `/wiki/` - Internal Company Wiki
Team documentation, processes, and company information

## Key Technologies

- **Backend**: Go 1.24+
- **Frontend**: Next.js (React), Astro
- **Infrastructure**: Terraform, Kubernetes, Helm
- **Cloud Platforms**: AWS (primary), Azure, GCP
- **Workflow Engine**: Temporal
- **Monitoring**: DataDog, OpenTelemetry
- **Databases**: PostgreSQL, ClickHouse

## Common Commands

Based on the repository structure, these commands are likely useful:

### Development
```bash
# CLI development
cd bins/cli && go run main.go

# Dashboard UI development  
cd services/dashboard-ui && npm run dev

# Wiki development
cd services/wiki && npm run dev
```

### Infrastructure
```bash
# Using nuonctl (administrative CLI)
./bins/nuonctl/scripts/[script-name]

# Terraform operations
cd infra/[module] && terraform plan
```

### Testing
```bash
# Go tests
go test ./...

# Frontend tests
cd services/dashboard-ui && npm test
```

### Code Generation
```bash
# IMPORTANT: Run this after making changes to Go types, temporal activities, 
# or any other code that has generated counterparts
./run-nuonctl.sh scripts reset-generated-code
```

**When to run code generation:**
- After adding/modifying Go struct types (especially in `/internal/app/`)
- After adding `@temporal-gen` annotations to functions
- After changing API endpoint definitions with swagger annotations
- When you see compilation errors about missing generated files
- When generated files (`.activity_gen.go`, `.workflow_gen.go`, swagger docs) are out of sync

## Getting Started

1. **Prerequisites**: Go 1.24+, Node.js, Docker, Terraform, kubectl
2. **Authentication**: Set up cloud credentials (AWS/Azure)  
3. **Local Development**: Check service-specific README files in `/services/` and `/bins/`
4. **Documentation**: Visit the `/docs/` directory or internal wiki

## Go Development Best Practices

When working with Go code in this repository, agents should follow these practices:

### Code Formatting
- **Always run `go fmt` after editing Go files** to ensure consistent formatting
- This prevents formatting inconsistencies and maintains code quality
- Example workflow:
  ```bash
  # After making changes to a Go file
  go fmt ./path/to/file.go
  
  # Or format entire directory
  go fmt ./services/ctl-api/...
  ```

### Code Quality
- Follow existing code patterns and conventions in each service
- Use proper error handling with meaningful error messages
- Add appropriate logging with structured fields
- Ensure proper imports and avoid unused dependencies

### API Development
- Use proper Swagger annotations for all HTTP endpoints
- Include both `@Security APIKey` and `@Security OrgID` for authenticated endpoints
- Follow the established route patterns in each service

## Notes for Claude

- This is a complex enterprise platform with many interconnected services
- The main business logic is in Go, with TypeScript for UIs
- Heavy use of Kubernetes, Terraform, and cloud-native technologies
- The `/bins/nuonctl/scripts/` directory contains many operational scripts
- Infrastructure code is in `/infra/` with Terraform modules
- Example applications and templates are in `/seed/`
- **CRITICAL**: Always run `go fmt` after making any changes to Go files

## User Journey & Onboarding System

The Nuon platform implements a comprehensive guided onboarding system:

### Architecture Overview
- **Backend**: Journey tracking in `ctl-api` with JSONB-stored user journey steps
- **Frontend**: Contextual modals in `dashboard-ui` that guide users through key actions
- **Integration**: Real-time journey updates via AccountProvider polling

### Journey Flow
1. **Account Creation** → User signs up
2. **Organization Creation** → User creates first org  
3. **App Configuration** → User runs `nuon apps sync` → Navigate to app page
4. **Install Creation** → User creates first install → Complete onboarding

### Key Implementation Details
- Journey steps store entity IDs (app_id, install_id) for navigation
- Modal system prevents infinite reopen loops via dismissal tracking  
- Cross-service journey updates via dependency injection
- Non-blocking: Journey failures never break core functionality

### Files to Reference
- **Backend**: `services/ctl-api/internal/app/accounts/helpers/update_user_journey_step.go`
- **Frontend**: `services/dashboard-ui/src/components/Apps/*Modal.tsx`
- **Data Structure**: `services/ctl-api/internal/app/user_journey.go`

## Account & Organization Permission System

Nuon uses a sophisticated multi-tenant RBAC (Role-Based Access Control) system for managing access to organizations.

### Account Types & Creation Flows

**Account Types** (`internal/app/account.go`):
- `AccountTypeAuth0` - Regular users (external customers)
- `AccountTypeService` - Service accounts for automation
- `AccountTypeCanary` - Internal testing accounts
- `AccountTypeIntegration` - Integration testing accounts

**Account Creation Paths** (`internal/middlewares/auth/account_token.go:71-104`):

1. **Self-Signup Flow**:
   - No pending `OrgInvite` found for email
   - Gets `DefaultEvaluationJourneyWithAutoOrg()` with user journey tracking
   - **Automatically creates trial org** with pattern `${email}-trial`
   - User becomes org admin immediately
   - Skips manual org creation step in dashboard

2. **Invite Flow**:
   - Pending `OrgInvite` exists for email
   - Gets `NoUserJourneys()` (no guided onboarding)
   - Invite acceptance grants specific org access via existing role assignment

### Permission Architecture (Three-Layer RBAC)

**1. Accounts** - Individual users or service accounts
**2. Roles** - Permission containers with specific purposes
**3. Policies** - Actual permission sets attached to roles

### Role System

**Standard Org Roles** (`internal/pkg/authz/create_org_roles.go`):
- `RoleTypeOrgAdmin` - Full organization administration
- `RoleTypeInstaller` - Install management permissions
- `RoleTypeRunner` - Runner execution permissions

Each role gets associated policies with permissions stored in PostgreSQL HSTORE format.

### How Accounts Access Organizations

**Account → Org Access Flow**:

1. **Org Role Creation** (`authzClient.CreateOrgRoles(ctx, orgID)`):
   - Creates standard roles for the organization
   - Each role gets policies with appropriate permissions
   - Requires account context for audit trail (`CreatedByID`)

2. **Account Role Assignment** (`authzClient.AddAccountOrgRole(ctx, roleType, orgID, accountID)`):
   - Creates `AccountRole` junction table entries
   - Links specific accounts to specific org roles
   - Uses conflict resolution to prevent duplicates

3. **Permission Resolution** (`internal/app/account.go:57-85`):
   - Account's `AfterQuery` hook aggregates permissions from all roles
   - Builds `OrgIDs` array from accessible organizations
   - Creates unified `AllPermissions` set for authorization

### Context & Audit Requirements

**CreatedByID Pattern**:
All major entities require audit tracking:
- Org, Role, Policy models have `CreatedByID` fields
- `BeforeCreate` hooks automatically populate from context
- **Critical**: Must set account context before operations:
  ```go
  ctx = cctx.SetAccountContext(ctx, account)
  ```

### Auto-Org Creation Implementation

**New Self-Signup Flow** (implemented):
- `CreateAccountWithAutoOrg()` creates account + trial org atomically
- Sets proper account context for org creation hooks
- Creates org roles and assigns user as admin
- User journey reflects completed org creation step

**Key Files**:
- `internal/pkg/account/create.go` - Account creation with auto org
- `internal/middlewares/auth/account_token.go` - Auth flow logic
- `internal/app/user_journey.go` - Journey step definitions
- `internal/pkg/authz/` - Role and permission management

### User Journey Integration

**Updated Journey Flow**:
1. **Account Created** → Self-signup account created
2. **Org Created** → Trial org automatically created (marked complete)
3. **App Created** → User runs `nuon apps sync`
4. **Install Created** → User creates first install

**Journey Helpers Pattern**:
- Cross-domain operations use helpers (e.g., `accountsHelpers.UpdateUserJourneyStep`)
- Helpers injected via FX dependency injection
- Non-blocking: Journey failures never break core operations

### Important Implementation Notes

- **Transaction Safety**: Auto org creation uses database transactions for atomicity
- **Error Handling**: Role creation failures return detailed error messages
- **Context Propagation**: Account context must be set for all authz operations
- **Invite Preservation**: Existing invite flow unchanged - only affects self-signup
- **Journey Tracking**: Auto-created orgs marked as completed in user journey

## CLAUDE.md Context Files

This monorepo contains 14 CLAUDE.md files that provide component-specific context and instructions for AI assistants. These files contain critical domain knowledge, development patterns, and service-specific guidance.

### CLAUDE.md File Locations

**Root Level:**
- `/CLAUDE.md` - Main project instructions (references this AGENTS.md file)

**Binary Tools (`/bins/`):**
- `/bins/cli/CLAUDE.md` - Public CLI tool (`nuon` command)
- `/bins/nuonctl/CLAUDE.md` - Internal CLI with operational scripts
- `/bins/runner/CLAUDE.md` - Deployment execution binary

**Services (`/services/`):**
- `/services/ctl-api/CLAUDE.md` - Control API service (Go)
- `/services/dashboard-ui/CLAUDE.md` - Main dashboard UI (Next.js/React)
- `/services/e2e/CLAUDE.md` - End-to-end testing service
- `/services/orgs-api/CLAUDE.md` - Organizations API (DEPRECATED)
- `/services/website-v2/CLAUDE.md` - Marketing website (Astro)
- `/services/website/CLAUDE.md` - Legacy website (Astro)
- `/services/wiki/CLAUDE.md` - Internal documentation site
- `/services/workers-canary/CLAUDE.md` - Canary testing service
- `/services/workers-executors/CLAUDE.md` - Core execution engine (DEPRECATED)
- `/services/workers-infra-tests/CLAUDE.md` - Infrastructure testing service

### Instructions for AI Assistants

**CRITICAL: Session Context Loading**

When starting any new session to work on this monorepo, AI assistants should:

1. **Always load the root context files first:**
   ```
   Read /CLAUDE.md (main project instructions)
   Read /AGENTS.md (this comprehensive overview)
   ```

2. **Load component-specific context based on the task:**
   - If working on the CLI: Read `/bins/cli/CLAUDE.md`
   - If working on the API: Read `/services/ctl-api/CLAUDE.md`
   - If working on the dashboard: Read `/services/dashboard-ui/CLAUDE.md`
   - If working across multiple services: Read all relevant CLAUDE.md files

3. **Use globbing to discover all context files:**
   ```bash
   # Find all CLAUDE.md files in the monorepo
   glob pattern: **/CLAUDE.md
   ```

4. **Load context systematically:**
   - Root context provides architectural overview and common patterns
   - Component-specific context provides detailed implementation guidance
   - Each CLAUDE.md file contains domain-specific knowledge not found elsewhere

### Benefits of Hierarchical Context

- **Component Isolation**: Each service/binary has specific development patterns and constraints
- **Historical Context**: Deprecated services maintain their context for reference
- **Specialized Knowledge**: Domain-specific implementation details and gotchas
- **Development Efficiency**: Reduces discovery time for service-specific conventions

### Maintenance Guidelines

- Keep CLAUDE.md files updated as services evolve
- Document major architectural changes in relevant component files
- Ensure root AGENTS.md reflects current monorepo structure
- Archive context files in `/graveyard/` when services are fully deprecated

## Project Status

Main branch: `main`
Repository is clean with recent commits related to CLI improvements and authentication.
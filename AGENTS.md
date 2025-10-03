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

## Running the Complete Stack Locally

The Nuon monorepo provides a sophisticated orchestration system for running the entire platform locally, including all microservices and their dependencies.

### Quick Start: Full Stack

**Run all services (RECOMMENDED):**
```bash
# First regenerate all code (Go types, temporal activities, swagger docs)
./run-nuonctl.sh scripts reset-generated-code

# Start infrastructure dependencies
./run-nuonctl.sh scripts exec reset-dependencies

# Then start all application services - SIMPLIFIED COMMAND
./run-nuonctl.sh dev --dev-all
```

**Alternative service-specific orchestration (if needed):**
```bash
./run-nuonctl.sh services dev --dev-all --skip cli,runner
```

> **⚠️ Important**: The simplified `dev --dev-all` command is more reliable than the service-specific `services dev --dev-all` variant. It properly handles service startup orchestration and dependencies, especially for the ctl-api's critical code generation phase.

### Complete Architecture

When running the full stack, you get:

**Application Services** (via `--dev-all`):
- **`ctl-api`** - Core backend API (ports 8081, 8082, 8083)
- **`dashboard-ui`** - Web frontend (port 4000)
- **`cli`** - Public CLI tool
- **`runner`** - Deployment execution engine
- **`wiki`** - Internal documentation (port 4321)

**Infrastructure Dependencies** (via Docker Compose):
- **PostgreSQL** - Main database with multiple databases (`ctl_api`, `temporal`, `temporal_visibility`)
- **ClickHouse Cluster** - Replicated analytics database (2 nodes + 3 keepers)
- **Temporal** - Workflow engine with web UI
- **PGAdmin** - Database management UI
- **Development Tools** - Registry, profiling, tunneling

### Step-by-Step Setup

1. **Regenerate Code:**
   ```bash
   ./run-nuonctl.sh scripts reset-generated-code
   ```
   This ensures all Go types, Temporal activities, and Swagger documentation are up-to-date.

2. **Start Infrastructure Dependencies:**
   ```bash
   ./run-nuonctl.sh scripts exec reset-dependencies
   ```
   This runs `docker-compose up` with all required databases and supporting services.

3. **Start Application Services:**
   ```bash
   # RECOMMENDED: Simplified orchestration
   ./run-nuonctl.sh dev --dev-all

   # Alternative: Service-specific orchestration
   ./run-nuonctl.sh services dev --dev-all --skip cli,runner
   ```

4. **Verify Services:**
   - Dashboard UI: http://localhost:4000
   - API Endpoints: http://localhost:8081 (public), http://localhost:8082 (admin)
   - Temporal UI: http://localhost:8233
   - Database UIs: http://localhost:8888 (PGAdmin), http://localhost:5521 (ClickHouse)

### Port Map for Full Stack

```
Frontend Services:
- Dashboard UI:     http://localhost:4000
- Wiki:             http://localhost:4321

Backend Services:
- CTL-API Public:   http://localhost:8081
- CTL-API Admin:    http://localhost:8082
- CTL-API Runner:   http://localhost:8083

Infrastructure:
- PostgreSQL:       localhost:5432
- ClickHouse:       localhost:9000 (native), localhost:8123 (HTTP)
- Temporal:         localhost:7233
- Temporal UI:      http://localhost:8233
- PGAdmin:          http://localhost:8888
- ClickHouse UI:    http://localhost:5521
```

### Common Development Patterns

**Backend-focused development:**
```bash
./run-nuonctl.sh services dev --dev ctl-api
```

**Frontend-focused development:**
```bash
./run-nuonctl.sh services dev --dev dashboard-ui,ctl-api
```

**Debug mode:**
```bash
./run-nuonctl.sh services dev --dev-all --debug
```

### Prerequisites for Full Stack

1. **Docker/Podman** - For infrastructure services
2. **Go 1.24+** - For Go services
3. **Node.js** - For frontend services
4. **AWS Credentials** - Services assume support role for real AWS access
5. **Auth0 Configuration** - For dashboard authentication

### Troubleshooting Full Stack

**Command Selection Issues:**
- **Use simplified command**: `./run-nuonctl.sh dev --dev-all` is more reliable than `services dev --dev-all`
- **CTL-API startup problems**: The simplified command properly handles ctl-api's code generation dependencies
- **Service orchestration**: Simplified command manages startup sequence and service dependencies automatically

**If services fail to start:**
- Check infrastructure: `docker ps`
- Verify ports: `lsof -i :4000,8081,5432`
- Check AWS access: `aws sts get-caller-identity`
- Run with debug: `--debug` flag

**Performance optimization:**
- Skip unused services: `--skip cli,runner`
- Disable file watching: `--watch=false`
- Use user overrides in `~/.nuonctl-env.yml`

## Getting Started

1. **Prerequisites**: Go 1.24+, Node.js, Docker, Terraform, kubectl
2. **Authentication**: Set up cloud credentials (AWS/Azure)
3. **Local Development**: Start with running the complete stack above
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

## Local Development with nuonctl

The Nuon monorepo uses a sophisticated development orchestration system built around the `nuonctl` CLI tool. Understanding this system is crucial for effective development and configuration management.

### Development Command Structure

**Primary Development Command:**
```bash
./run-nuonctl.sh services dev --dev <service-name>
```

**Common Usage Examples:**
```bash
# Run dashboard UI locally
./run-nuonctl.sh services dev --dev dashboard-ui

# Run multiple services
./run-nuonctl.sh services dev --dev dashboard-ui,ctl-api

# Run all services in development mode
./run-nuonctl.sh services dev --dev-all

# Skip specific services
./run-nuonctl.sh services dev --dev-all --skip cli,runner
```

### Configuration Architecture

Nuon uses a **two-tier configuration system** that provides both team consistency and individual flexibility:

#### 1. Service Base Configuration (`service.yml`)

Each service contains a `service.yml` file with default development settings:

```yaml
# /services/dashboard-ui/service.yml
env:                    # Local development environment
  NEXT_PUBLIC_API_URL: 'http://localhost:8081'
  NUON_WORKFLOWS: true
  FEATURE_FLAG_NAME: false

docker_env:             # Docker container environment
  NEXT_PUBLIC_API_URL: 'http://host.nuon.dev:8081'
  FEATURE_FLAG_NAME: false

local_cmds:             # Commands executed for local development
  - npm run dev

local_startup_cmds: []  # Commands run before service startup
```

**Key Sections:**
- `env:` - Environment variables for local development
- `docker_env:` - Environment variables when running in Docker containers
- `local_cmds:` - Commands executed to start the service locally
- `tests:` - Test configurations and Docker targets

#### 2. User Override System (`~/.nuonctl-env.yml`)

Individual developers can override any service configuration:

```yaml
# ~/.nuonctl-env.yml
dashboard-ui:
  NEXT_PUBLIC_ENABLE_FULL_SCREEN_ONBOARDING: true
  NUON_WORKFLOWS: false
  CUSTOM_DEBUG_FLAG: true

ctl-api:
  LOG_LEVEL: DEBUG
  CHAOS_RATE: 1

runner:
  DISABLE_INSTALL_RUNNER: true
```

**Override Behavior:**
- User overrides take **precedence** over service.yml values
- Only specified values are overridden; others use service.yml defaults
- Overrides are per-service using the service directory name as the key

### Configuration Merge Process

When `nuonctl services dev` runs:

1. **Load Base Config**: Reads `/services/<service>/service.yml`
2. **Load User Overrides**: Reads `~/.nuonctl-env.yml` (if exists)
3. **Merge Configurations**: User overrides take precedence
4. **Export Environment**: Combined environment variables are exported
5. **Execute Commands**: Runs `local_cmds` with merged environment

### Adding New Environment Variables

**For Service Defaults (Team-Wide):**

Edit the service's `service.yml` file in both sections:
```yaml
env:
  NEW_FEATURE_FLAG: false        # Local development default

docker_env:
  NEW_FEATURE_FLAG: false        # Docker container default
```

**For Individual Testing:**

Add to your personal `~/.nuonctl-env.yml`:
```yaml
service-name:
  NEW_FEATURE_FLAG: true         # Personal override
```

### Service Management Commands

**Service Discovery:**
```bash
# List all available services
./run-nuonctl.sh services --help

# Available services: ctl-api, dashboard-ui, cli, runner, wiki, docs, website
```

**Development Modes:**
```bash
# Local development (runs npm/go commands directly)
./run-nuonctl.sh services dev --dev dashboard-ui --dev-type local

# Docker development (runs services in containers)
./run-nuonctl.sh services dev --dev dashboard-ui --dev-type docker
```

**Service Control:**
```bash
# Watch for file changes and auto-restart (default: enabled)
./run-nuonctl.sh services dev --dev dashboard-ui --watch

# Disable file watching
./run-nuonctl.sh services dev --dev dashboard-ui --watch=false

# Pull latest images instead of building locally
./run-nuonctl.sh services dev --dev dashboard-ui --pull

# Build images locally instead of pulling
./run-nuonctl.sh services dev --dev dashboard-ui --pull=false
```

### Environment Variable Debugging

**Enable Debug Mode:**
```bash
./run-nuonctl.sh services dev --dev dashboard-ui --debug
```

This will show:
- All environment variables being set
- Configuration merge results
- Commands being executed
- Override sources and values

### Common Development Patterns

#### Feature Flag Development
```yaml
# ~/.nuonctl-env.yml
dashboard-ui:
  NEXT_PUBLIC_NEW_FEATURE: true
  NUON_EXPERIMENTAL_MODE: true
```

#### API Endpoint Switching
```yaml
# ~/.nuonctl-env.yml
dashboard-ui:
  NEXT_PUBLIC_API_URL: 'https://api.nuon-stage.co'  # Point to staging
  NUON_API_URL: 'https://api.nuon-stage.co'
```

#### Multi-Service Development
```bash
# Run API and UI together with custom configs
./run-nuonctl.sh services dev --dev ctl-api,dashboard-ui
```

### Troubleshooting Development Issues

**Configuration Problems:**
1. Run with `--debug` to see environment variable merge results
2. Check `~/.nuonctl-env.yml` for conflicting overrides
3. Verify service.yml syntax with a YAML validator

**Service Startup Issues:**
1. Check if ports are already in use (default: UI=4000, API=8081)
2. Verify dependencies are running (e.g., database, temporal)
3. Look for authentication issues (AWS SSO, Auth0)

**File Change Detection:**
1. NPM projects have file watching disabled by default
2. Go services automatically restart on file changes
3. Use `--watch=false` if auto-restart causes issues

### Integration with Code Generation

**Important**: Many services have code generation steps that must be run when making changes:

```bash
# Reset and regenerate all code
./run-nuonctl.sh scripts reset-generated-code

# This is required after:
# - Adding/modifying Go struct types
# - Adding @temporal-gen annotations
# - Changing API endpoint definitions with swagger annotations
```

### Best Practices for Agents

1. **Always check service.yml first** when working with environment variables
2. **Use ~/.nuonctl-env.yml for testing** rather than modifying service.yml
3. **Run with --debug** when troubleshooting configuration issues
4. **Remember to run code generation** after significant changes
5. **Check for existing overrides** in ~/.nuonctl-env.yml when debugging

This development system provides powerful flexibility while maintaining team consistency. It allows for rapid iteration and testing without requiring changes to shared configuration files.

## Project Status

Main branch: `main`
Repository is clean with recent commits related to CLI improvements and authentication.
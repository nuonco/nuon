# Nuon Monorepo Overview

This is a comprehensive monorepo for Nuon, a BYOC (Bring Your Own Cloud) platform that helps software vendors deploy
applications to their customers' cloud accounts.

## Repository Structure

This monorepo is written primarily in **Go** (main module: `github.com/nuonco/nuon`) with several TypeScript/JavaScript
projects for UI components.

### Core Directories

#### `/bins/` - Command Line Tools & Executables

Binaries compiled and run as executables (not deployed as Kubernetes services):

- **`cli/`** - Public-facing CLI tool for developers (`nuon` command)
- **`nuonctl/`** - Internal CLI with 80+ operational scripts and dev tooling
- **`runner/`** - Deployment execution binary (runs in customer K8s + VMs)

#### `/services/` - Web Services & Applications

- **`ctl-api/`** - Control API service (Go)
- **`dashboard-ui/`** - Main dashboard UI: the running app is a client-side SPA served by a Go server from the `client/`
  directory (not the `src/` Next.js app)
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

- **Backend**: Go 1.25+
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

Use `go generate` to regenerate code after making changes:

```bash
# Temporal activities (ctl-api)
go generate ./services/ctl-api/...

# Temporal workflow types (executors)
go generate ./pkg/workflows/types/executors/...

# SDKs (nuon-go and nuon-runner-go)
go generate ./sdks/nuon-go/...
go generate ./sdks/nuon-runner-go/...

# Regenerate all code in the monorepo
go generate ./...
```

**When to run code generation:**

- After adding/modifying Go struct types (especially in `/internal/app/`)
- After adding `@temporal-gen` annotations to functions
- After changing API endpoint definitions with swagger annotations
- When you see compilation errors about missing generated files
- When generated files (`.activity_gen.go`, `.workflow_gen.go`, swagger docs) are out of sync

**Key `go:generate` locations:**

- `services/ctl-api/main.go` - API types and swagger docs
- `pkg/workflows/types/executors/gen.go` - Temporal activity/workflow generation
- `sdks/nuon-go/client.go` - Go SDK client generation
- `sdks/nuon-runner-go/client.go` - Runner SDK client generation

## Running Services Locally

Ask the user how to start the services to test your changes.

### Expected URLs

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
```

### Local Debugging via Docker

Use Docker to inspect Temporal, Postgres, and ClickHouse when services are running locally.

```bash
# Temporal CLI (components namespace)
temporal workflow list --namespace components
temporal workflow show --namespace components --workflow-id <workflow_id> --run-id <run_id> --output json

# Postgres (ctl-api)
docker exec -i mono-postgres-1 psql -U ctl_api -d ctl_api

# Example Postgres query
docker exec -i mono-postgres-1 psql -U ctl_api -d ctl_api \
  -c "select id, status, status_description from runner_jobs order by created_at desc limit 5;"

# ClickHouse (ctl-api logs)
docker exec -i clickhouse-01 clickhouse-client \
  --query "select timestamp, severity_text, body from ctl_api.otel_log_records order by timestamp desc limit 50"
```

## Getting Started

1. **Prerequisites**: Go 1.25+, Node.js, Docker, Terraform, kubectl
2. **Authentication**: Set up cloud credentials (AWS/Azure)
3. **Local Development**: Start with running the complete stack above
4. **Documentation**: Visit the `/docs/` directory or internal wiki

## Go Development Best Practices

When working with Go code in this repository, agents should follow these practices:

### Code Formatting

- **CRITICAL: Always run formatting tools after making ANY Go code changes**
- **Use `gofmt` and `goimports` on directories/packages, not individual files**
- Run these commands automatically after each change to prevent import and formatting issues
- **Do NOT manually manage imports** - let `goimports` handle it
- This prevents formatting inconsistencies and maintains code quality

**Recommended workflow after making Go changes:**

```bash
# Format and fix imports for entire package/directory (PREFERRED)
gofmt -w ./services/ctl-api/internal/app/installs/worker/plan/
goimports -w ./services/ctl-api/internal/app/installs/worker/plan/

# Or format recursively from a parent directory
gofmt -w ./services/ctl-api/...
goimports -w ./services/ctl-api/...

# For pkg/ changes
gofmt -w ./pkg/plans/types/
goimports -w ./pkg/plans/types/
```

**Why this matters:**

- `gofmt` ensures consistent code style across the codebase
- `goimports` automatically adds missing imports and removes unused ones
- Running on directories catches all changed files in one pass
- Prevents compilation errors from missing or incorrectly ordered imports
- Avoids the need to manually add/remove import statements

### Code Quality

- Follow existing code patterns and conventions in each service
- Use proper error handling with meaningful error messages
- Add appropriate logging with structured fields
- Ensure proper imports and avoid unused dependencies

### GORM Query Conventions

**Always use struct-based `Where` clauses instead of raw SQL strings.** This applies to all GORM queries throughout the
codebase.

```go
// ✅ CORRECT - Struct-based Where clause
db.Where(app.QueueSignal{
    OwnerID:   stepID,
    OwnerType: (&app.WorkflowStep{}).TableName(),
}).First(&qs)

// ❌ WRONG - Raw SQL string Where clause
db.Where("owner_id = ? AND owner_type = ?", stepID, "install_workflow_steps").First(&qs)
```

**Why:**

- Type-safe: compiler catches field name typos
- Refactor-safe: field renames propagate automatically
- Consistent: matches the established codebase pattern

**Note:** Use `(&app.Model{}).TableName()` (pointer receiver) when the model's `TableName()` method has a pointer
receiver.

### Logging

**Never use `fmt.Println` for logging.** See [conventions/logging.md](/conventions/logging.md) for full guidelines.

| Component                 | Logger                  |
| ------------------------- | ----------------------- |
| `pkg/` with logger access | `*zap.Logger` passed in |
| `pkg/` init functions     | Standard `log` package  |

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

**1. Accounts** - Individual users or service accounts **2. Roles** - Permission containers with specific purposes **3.
Policies** - Actual permission sets attached to roles

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

**CreatedByID Pattern**: All major entities require audit tracking:

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

This monorepo contains 14 CLAUDE.md files that provide component-specific context and instructions for AI assistants.
These files contain critical domain knowledge, development patterns, and service-specific guidance.

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

## Verifying OCI Artifacts from Local Development

When running locally, build artifacts (OCI images) are pushed to the **orgs-stage** AWS account ECR registry. This
section documents how to verify artifacts without running the install runner.

### ECR Configuration for Local Development

- **AWS Account**: `766121324316` (orgs-stage) - _loaded from config, not hardcoded_
- **Region**: `us-west-2`
- **Repository Pattern**: `<org_id>/<app_id>`
- **Tag**: Build ID (e.g., `bldq7fplr1up5atx5zpxotbabm`)

The configuration is loaded from the Kubernetes ConfigMap `ctl-api-stage` in the stage cluster. To check current values:

```bash
kubectl get configmap ctl-api-stage -n ctl-api -o jsonpath='{.data}' | jq -r 'to_entries[] | select(.key | contains("MANAGEMENT")) | "\(.key): \(.value)"'
```

Key config values:

- `MANAGEMENT_ACCOUNT_ID`: AWS account ID for ECR (currently `766121324316`)
- `MANAGEMENT_ECR_REGISTRY_ID`: ECR registry ID (currently `766121324316`)
- `MANAGEMENT_IAM_ROLE_ARN`: Role to assume for ECR access

### Verifying Build Artifacts

**1. Login to orgs-stage ECR:**

```bash
# Using orgs-stage.NuonAdmin profile directly
aws ecr get-login-password --region us-west-2 --profile orgs-stage.NuonAdmin | \
  docker login --username AWS --password-stdin 766121324316.dkr.ecr.us-west-2.amazonaws.com
```

**2. List images in a repository:**

```bash
aws ecr list-images --region us-west-2 --profile orgs-stage.NuonAdmin \
  --repository-name "<org_id>/<app_id>"
```

**3. Pull and inspect artifact with ORAS:**

```bash
# Pull the artifact
oras pull 766121324316.dkr.ecr.us-west-2.amazonaws.com/<org_id>/<app_id>:<build_id> \
  -o /tmp/artifact-verify

# View contents
ls /tmp/artifact-verify
cat /tmp/artifact-verify/manifest.yaml  # For Kubernetes manifest builds
```

### Artifact Contents by Component Type

| Component Type      | Artifact Contents                                                     |
| ------------------- | --------------------------------------------------------------------- |
| Kubernetes Manifest | `manifest.yaml` - Rendered YAML (kustomize output or inline manifest) |
| Terraform Module    | `*.tf` files - Terraform configuration files                          |
| Helm Chart          | Chart archive with templates                                          |
| Docker Build        | Container image layers                                                |

### Example: Verifying a Kubernetes Manifest Build

```bash
# Given:
# - org_id: orgrok933tcyzji01s7us3aeo3
# - app_id: app98e2wpzdxwoey393edtqj45
# - build_id: bldq7fplr1up5atx5zpxotbabm

# Pull the artifact
oras pull 766121324316.dkr.ecr.us-west-2.amazonaws.com/orgrok933tcyzji01s7us3aeo3/app98e2wpzdxwoey393edtqj45:bldq7fplr1up5atx5zpxotbabm -o /tmp/verify

# View the rendered manifest
cat /tmp/verify/manifest.yaml
```

### Troubleshooting

**Repository not found:**

- Ensure you're in the correct AWS account (766121324316, not the stage account 676549690856)
- Check that the app creation workflow completed successfully (ECR repo is created during app provisioning)

**Authentication errors:**

- Refresh your AWS SSO session: `aws sso login --profile stage.NuonAdmin`
- Re-assume the support role

**Empty repository:**

- Build workflow may have failed - check Temporal UI for workflow status
- Check runner logs for build errors

## Notification Surfaces — Signal → Webhook + Slack Workflow (IMPORTANT)

**IMPORTANT:** Whenever a new signal is created in `services/ctl-api/internal/app/**/signals/`, the work is NOT done at
the signal definition. The signal must be propagated through every notification surface the platform exposes.

### Required end-to-end workflow

For every new signal, follow this sequence — do not stop after step 1:

1. **Create + implement the webhook + Slack notification hook for the signal.**

   - Wire the signal into the webhook dispatcher at `services/ctl-api/internal/pkg/queue/signal/hooks/webhook.go`
     so emissions of the signal turn into outbound webhook deliveries.
   - Wire the signal into the Slack notification hook in the same `hooks/` package (and any matching renderer in
     `services/ctl-api/internal/app/slack/`) so emissions of the signal post into subscribed channels.
   - Both hooks consult the per-row `Match` (routing predicate) and `Interests` (event filter) before delivery —
     keep that contract intact.

2. **Prompt the user for filter specifics — `match` and `interests`.**

   - Ask the user explicitly: which `interests` resources/ops should this signal map to (installs / components /
     actions / workflows; ops like `deploy`, `destroy`, `plan`; outcomes like `completion` / `failures`; approval
     requests / responses), and how should `match` be interpreted for it (org-wide, specific entity IDs, or label
     selector).
   - Do NOT assume defaults — these decisions determine the dispatch-time `labels.EventTargets` and the
     `interests.Interests` shape, and getting them wrong silently mis-routes deliveries.

3. **Prompt the user to implement the new signal across every consumer surface — for BOTH webhooks and Slack.**

   The signal must be exposed in every place a user can configure or observe a notification rule. Confirm with the
   user before starting, then implement in:

   - **Dashboard UI** (`services/dashboard-ui/client/components/`):
     - `webhooks/CreateWebhook`, `webhooks/EditWebhook`, `webhooks/WebhooksTable`
     - `slack/CreateChannelSubscription`, `slack/EditChannelSubscription`, `slack/ChannelSubscriptionsTable`
     - `interests/` and `match/` pickers if the new signal introduces a new resource kind, op, or selector axis
   - **CLI** (`bins/cli/`):
     - The `nuon webhooks` and `nuon slack` (or equivalent notification) command groups — create / edit flows must
       accept the new signal in their `--interests` / `--match` flag parsing.
   - **Slack subscribe modal** (`services/ctl-api/internal/app/slack/service/subscribe_modal.go`):
     - The `/nuon subscribe` slash-command modal renders the same picker surface — keep its options in lockstep
       with the dashboard's `MatchPicker` / `InterestsPicker`.

### Why this matters

The webhook hook (`pkg/queue/signal/hooks/webhook.go`), the Slack hook (same package), the dashboard pickers
(`client/components/match/`, `client/components/interests/`), the CLI flag parsers, and the Slack slash-command modal
all share a single contract — `labels.SubscriptionMatch` + `interests.Interests`. Skipping any surface leaves users
unable to subscribe to the new signal even though the backend emits it, which presents as a silent regression rather
than an explicit error.

### Quick reference

| Layer                | Key file(s)                                                                    |
| -------------------- | ------------------------------------------------------------------------------ |
| Signal definition    | `services/ctl-api/internal/app/**/signals/v2/<signal>/signal.go`               |
| Webhook hook         | `services/ctl-api/internal/pkg/queue/signal/hooks/webhook.go`                  |
| Slack hook           | `services/ctl-api/internal/pkg/queue/signal/hooks/` (slack-side renderer)      |
| Match contract       | `pkg/labels/match.go` (Go) ↔ `services/dashboard-ui/client/components/match/types.ts` (TS) |
| Interests contract   | `services/ctl-api/internal/pkg/interests/` (Go) ↔ `services/dashboard-ui/client/components/interests/` (TS) |
| Dashboard webhooks   | `services/dashboard-ui/client/components/webhooks/`                            |
| Dashboard Slack      | `services/dashboard-ui/client/components/slack/`                               |
| CLI commands         | `bins/cli/internal/cmd/webhooks/`, `bins/cli/internal/cmd/slack/`              |
| Slack slash-command  | `services/ctl-api/internal/app/slack/service/subscribe_modal.go`               |

## Project Status

Main branch: `main` Repository is clean with recent commits related to CLI improvements and authentication.

# CTL-API Service

The **Control API (ctl-api)** is the core backend service of the Nuon platform, providing comprehensive APIs for
managing applications, components, installs, and infrastructure deployments.

## Service Overview

This is a Go-based microservice that serves as the primary API gateway for the Nuon platform. It provides three distinct
API surfaces:

- **Public API** - For external users and CLI tools
- **Runner API** - For Nuon runners executing deployments
- **Admin API** - For internal administrative operations

## Architecture

- **Language**: Go
- **Framework**: Gin HTTP framework with extensive middleware
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT-based with Auth0 integration
- **Workflow Engine**: Temporal for orchestrating complex operations
- **Metrics**: DataDog integration via tally
- **Documentation**: Auto-generated Swagger/OpenAPI specs

## Relationship to Other Services

- **Primary consumer**: `dashboard-ui` service (main frontend)
- **CLI integration**: Both `cli` and `nuonctl` binaries
- **Runner communication**: Communicates with `runner` binaries in customer infrastructure
- **Workflow orchestration**: Uses Temporal workers for background processing
- **Infrastructure**: Manages deployments via `workers-executors`

## Project Structure

### Core Files

- `main.go` - Application entry point
- `public.go` - Public API routes and handlers
- `admin.go` - Admin API routes and handlers
- `runner.go` - Runner API routes and handlers
- `service.yml` - Service configuration

### Key Directories

#### `/internal/app/` - Domain Models

Contains all database models and business logic:

- `account.go`, `org.go` - User and organization management
- `app*.go` - Application configuration and metadata
- `component*.go` - Component definitions and builds
- `install*.go` - Installation and deployment tracking
- `runner*.go` - Runner management and job execution
- `terraform_*.go` - Terraform state management
- `vcs_*.go` - Version control system integration

Each domain follows a consistent structure:
```
/internal/app/{domain}/
├── service/          # HTTP handlers and API endpoints
├── helpers/          # Shared business logic and utilities
├── worker/           # Temporal workflows and activities
└── signals/          # Event definitions
```

#### `/internal/pkg/` - Business Logic

- `api/` - API service definitions and middleware setup
- `account/` - Account management services
- `activities/` - Temporal activity implementations
- `authz/` - Authorization and permission handling

#### `/internal/middlewares/` - HTTP Middleware

- `auth/` - Authentication middleware
- `org/` - Organization context injection
- `metrics/` - Request metrics collection
- `cors/` - Cross-origin resource sharing
- `admin/` - Admin-only access controls

#### `/docs/` - API Documentation

- `public/` - Public API Swagger documentation
- `admin/` - Admin API documentation
- `runner/` - Runner API documentation Auto-generated from code annotations.

#### `/infra/` - Infrastructure as Code

Terraform configuration for deploying the ctl-api service:

- `rds.tf` - PostgreSQL database setup
- `service.tf` - ECS/Kubernetes service configuration
- `dns_management.tf` - Route 53 DNS setup

#### `/k8s/` - Kubernetes Deployment

Helm chart templates for Kubernetes deployment:

- `templates/` - K8s resource templates
- `values.yaml` - Default configuration values

## Key Features

### Multi-Tenant Architecture

- Organization-based isolation
- Role-based access control
- Account delegation for customer access

### Component Management

- Docker builds, Helm charts, Terraform modules
- Dependency tracking and build orchestration
- Release management and versioning

### Install Management

- Infrastructure provisioning and deployment
- Workflow orchestration with approvals
- State management and rollback capabilities

### Runner Integration

- Secure communication with customer infrastructure
- Job execution and status reporting
- Health monitoring and metrics collection

### Admin Operations

- Organization management and feature flags
- Infrastructure debugging and troubleshooting
- Bulk operations and data migration tools

## Helpers Pattern

The ctl-api uses a helpers pattern to share domain-specific business logic across services while maintaining clean separation of concerns.

### Structure

Each domain may have a `/helpers` directory containing:

- **`helpers.go`** - Main helpers struct with FX dependency injection
- **Individual helper files** - Specific functionality (e.g., `update_user_journey_step.go`)

### Usage Pattern

```go
// 1. Define helpers struct with dependencies
type Helpers struct {
    cfg *internal.Config
    db  *gorm.DB
    v   *validator.Validate
}

// 2. Register in FX dependency injection (cmd/cli.go)
fx.Provide(accountshelpers.New),

// 3. Inject into services that need the functionality
type Params struct {
    fx.In
    AccountsHelpers *accountshelpers.Helpers
    // ... other dependencies
}

// 4. Use helper methods in service handlers
func (s *service) CreateOrg(ctx *gin.Context) {
    // ... org creation logic ...
    
    // Use accounts helper for cross-domain functionality
    if err := s.accountsHelpers.UpdateUserJourneyStepForFirstOrg(ctx, acct.ID); err != nil {
        s.l.Warn("failed to update user journey", zap.Error(err))
    }
}
```

### Benefits

- **Cross-domain functionality** without direct service imports
- **Reusable business logic** across multiple services
- **Clean dependency management** via FX injection
- **Testable** - helpers can be mocked independently
- **Consistent patterns** - follows established conventions

### Examples

Current helpers implementations:
- `accounts/helpers` - User journey management
- `orgs/helpers` - Organization operations (hard delete, etc.)
- `runners/helpers` - Runner job management
- `components/helpers` - Component builds and dependencies
- `installs/helpers` - Installation workflows and validation

## Development

### Running Locally

```bash
cd services/ctl-api
go run main.go
```

### Key Commands

- `go run cmd/gen/main.go` - Generate API documentation
- `go run main.go worker` - Run Temporal worker
- `go run main.go admin` - Admin CLI operations

### API Development Best Practices

#### Adding New Endpoints

When adding new API endpoints, follow this process to ensure proper documentation generation:

1. **Create the endpoint handler** with proper Swagger annotations:
   ```go
   //	@ID						YourEndpointName
   //	@Summary				Brief description
   //	@Description			Detailed description
   //	@Tags					service_name
   //	@Accept					json
   //	@Produce				json
   //	@Security				APIKey
   //	@Security				OrgID
   //	@Param					param_name	path/query/body	type	required	"Description"
   //	@Success				200		{object}	ResponseType
   //	@Failure				400		{object}	stderr.ErrResponse
   //	@Router					/v1/your/endpoint [METHOD]
   ```

2. **Register the route** in the appropriate service file:
   ```go
   api.METHOD("/v1/your/endpoint", s.YourEndpointHandler)
   ```

3. **Documentation is auto-generated** - No manual regeneration needed:
   - The service automatically generates swagger docs on startup
   - Manual regeneration can cause issues and is not required

#### Swagger Documentation Issues

**Common Problem**: Missing markdown description files causing generation failures.

**Symptoms**:
- `ParseComment error: Unable to find markdown file` errors
- API fails to start with swagger parsing errors
- Documentation generation stops completely

**Solution**:
1. Check if referenced markdown files exist in `docs/public/descriptions/`
2. Create any missing `.md` files (they can be empty initially)
3. Restart the service - documentation will auto-generate

**Important Notes**:
- Generated files (`swagger.json`, `swagger.yaml`, `docs.go`) are **not tracked in git**
- Only markdown description files in `docs/public/descriptions/` are version controlled
- **DO NOT manually regenerate documentation** - the service handles this automatically
- The service auto-generates swagger docs on startup from code annotations
- Manual regeneration with `go run cmd/gen/main.go` can cause issues and should be avoided

### API Endpoints

- Public API: `/v1/*`
- Runner API: `/runner/*`
- Admin API: `/admin/*`
- Health checks: `/livez`, `/readyz`

## Configuration

Configuration is handled through:

- Environment variables
- YAML configuration files in `/infra/vars/`
- Service mesh configuration in `service.yml`

## Testing

Integration tests in `/internal/integration/` cover:

- API endpoint functionality
- Database operations
- Authentication and authorization
- Multi-tenant isolation

This service is the heart of the Nuon platform, orchestrating all deployment activities and providing the primary
interface for users, runners, and administrative operations.

## User Journey Tracking System

The ctl-api implements a comprehensive user journey tracking system for guided onboarding:

### Journey Step Structure
```go
type UserJourneyStep struct {
    Name      string `json:"name" gorm:"column:name"`
    Title     string `json:"title" gorm:"column:title"`  
    Complete  bool   `json:"complete" gorm:"column:complete;default:false"`
    AppID     string `json:"app_id,omitempty" gorm:"column:app_id"`        // For navigation
    InstallID string `json:"install_id,omitempty" gorm:"column:install_id"` // For navigation
}
```

### Journey Helper Pattern
Location: `internal/app/accounts/helpers/update_user_journey_step.go`

Pattern for adding new journey step completion:
```go
func (h *Helpers) UpdateUserJourneyStepForFirst[Entity](ctx context.Context, accountID, entityID string) error {
    // 1. Get account with journey data
    // 2. Find evaluation journey and specific step
    // 3. Only update if step is incomplete (first-time only)  
    // 4. Store entity ID for navigation
    // 5. Save with Select("user_journeys") for JSONB update
}
```

### Integration Pattern
In service endpoints after successful operations:
```go
user, err := cctx.AccountFromGinContext(ctx)
if err == nil {
    if err := s.accountsHelpers.UpdateJourneyStep(ctx, user.ID, entityID); err != nil {
        // CRITICAL: Log but don't fail the operation
        s.l.Warn("journey step update failed", zap.Error(err))
    }
}
```

### Current Journey Steps
- `account_created` - User signup complete
- `org_created` - First organization created  
- `app_created` - First app synced via `nuon apps sync` (stores app ID)
- `install_created` - First install created (stores install ID)

### Cross-Service Dependencies
Services needing journey updates must include:
```go
// In Params struct
AccountsHelpers *accountshelpers.Helpers

// In service struct  
accountsHelpers *accountshelpers.Helpers

// In constructor
accountsHelpers: params.AccountsHelpers,
```

### Current Integrations
- **App Config Sync**: `apps/service/create_app_config.go:74` - Updates `app_created` step
- **Install Creation**: `installs/service/create_install.go:112` - Updates `install_created` step

## HTTP Handler Patterns

The ctl-api follows a **domain-driven architecture** with clear separation of concerns for HTTP handlers and business logic.

### Architecture Overview

Each domain in `/internal/app/{domain}/` has:

- **`service/`** - HTTP handlers and private domain-specific methods
- **`helpers/`** - Methods shared across domains (cross-domain functionality only)  
- **`worker/`** - Temporal workflows and activities (optional)

### Business Logic Placement Rules

**Critical Rule**: Business logic placement depends on whether it needs to be shared across domains:

- **Domain-specific logic** → Private service methods within the same service
- **Cross-domain shared logic** → Helpers package  
- **HTTP concerns only** → Handler methods

### Handler Pattern Examples

#### ✅ Good: Handler → Private Service Method (Domain-Specific Logic)

Use this pattern when business logic is specific to the domain and doesn't need to be shared.

**File**: `accounts/service/get_current_account.go`
```go
// Handler focuses on HTTP concerns
func (s *service) GetCurrentAccount(ctx *gin.Context) {
    account, err := cctx.AccountFromContext(ctx)
    if err != nil {
        ctx.Error(err)
        return
    }

    // Delegate to private service method for domain-specific logic
    fullAccount, err := s.getAccount(ctx, account.ID)
    if err != nil {
        ctx.Error(err)
        return
    }

    ctx.JSON(http.StatusOK, fullAccount)
}

// Private method contains domain-specific business logic
func (s *service) getAccount(ctx *gin.Context, accountID string) (*app.Account, error) {
    var account app.Account
    res := s.db.WithContext(ctx).
        Preload("Roles").
        Preload("Roles.Policies").
        Preload("Roles.Org").
        Where("id = ?", accountID).
        First(&account)
    
    // Domain-specific business logic here
    return &account, res.Error
}
```

#### ✅ Good: Handler → Helpers (Cross-Domain Shared Logic)

Use helpers when multiple domains need the same functionality.

**File**: `orgs/service/create_org.go`
```go
func (s *service) CreateOrg(ctx *gin.Context) {
    // ... org creation logic ...

    // Use helpers for cross-domain functionality
    if err := s.accountsHelpers.UpdateUserJourneyStepForFirstOrg(ctx, acct.ID); err != nil {
        s.l.Warn("failed to update user journey for first org")
    }
}
```

**Helper**: `accounts/helpers/update_user_journey_step.go`
```go
// This helper is used by orgs, apps, and installs services
func (h *Helpers) UpdateUserJourneyStepForFirstOrg(ctx context.Context, accountID string) error {
    // Cross-domain logic for user journey updates
    // Used by multiple domains: orgs, apps, installs
}
```

**Cross-domain usage evidence:**
- `orgs/service/create_org.go:74`
- `apps/service/create_app_config.go:74` 
- `installs/service/create_install.go:112`

#### ❌ Bad: Inline Business Logic (Needs Refactoring)

Avoid embedding complex business logic directly in handlers.

```go
// BAD: Business logic mixed with HTTP concerns
func (s *service) ComplexHandler(ctx *gin.Context) {
    var req SomeRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.Error(err)
        return
    }

    // 50+ lines of complex business logic inline...
    // Database queries, validation, processing...
    // This should be extracted!
}
```

**Should be refactored to:**
```go
// GOOD: Clean separation
func (s *service) ComplexHandler(ctx *gin.Context) {
    var req SomeRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.Error(err)
        return
    }

    // Delegate to private service method (domain-specific logic)
    result, err := s.processComplexOperation(ctx, &req)
    if err != nil {
        ctx.Error(err)
        return
    }

    ctx.JSON(http.StatusOK, result)
}

// Private method contains the extracted business logic
func (s *service) processComplexOperation(ctx context.Context, req *SomeRequest) (*Result, error) {
    // All the complex business logic extracted here
}
```

### Handler Structure Template

```go
func (s *service) HandlerName(ctx *gin.Context) {
    // 1. Extract context data (org, user, params)
    org, err := cctx.OrgFromGinContext(ctx)
    if err != nil {
        ctx.Error(err)
        return
    }
    
    // 2. Parse and validate request
    var req RequestType
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.Error(fmt.Errorf("unable to parse request: %w", err))
        return
    }
    if err := req.Validate(s.v); err != nil {
        ctx.Error(fmt.Errorf("invalid request: %w", err))
        return
    }

    // 3. Delegate business logic appropriately
    if needsCrossDomainLogic {
        result, err := s.domainHelpers.CrossDomainOperation(ctx, &req)
    } else {
        result, err := s.domainSpecificOperation(ctx, &req)
    }
    
    if err != nil {
        ctx.Error(fmt.Errorf("operation failed: %w", err))
        return
    }

    // 4. Return response
    ctx.JSON(http.StatusCreated, result)
}
```

### When to Use Helpers vs Private Methods

**Use Helpers when:**
- Multiple domains need the same functionality
- Logic involves updating entities from different domains
- Common validation or processing logic is shared
- Examples: User journey updates, cross-domain validation, shared calculations

**Use Private Service Methods when:**
- Logic is specific to the current domain
- Database operations on domain entities
- Domain-specific business rules
- Internal processing that doesn't need sharing

### Best Practices

1. **Keep handlers minimal** - Focus on HTTP request/response handling
2. **Extract business logic** - Use private methods for domain logic, helpers for cross-domain logic
3. **Maintain clear boundaries** - Don't mix HTTP concerns with business logic
4. **Follow naming conventions**:
   - Handlers: `PascalCase` matching HTTP route (`CreateInstall`, `GetOrgs`)
   - Private methods: `camelCase` describing operation (`getInstall`, `validateInputs`)
   - Helper methods: `PascalCase` describing business operation (`UpdateUserJourneyStep`)

This pattern ensures clean separation of concerns, promotes code reuse where appropriate, and maintains clear domain boundaries.

## Account & Organization Permission System (RBAC)

The ctl-api implements a sophisticated multi-tenant Role-Based Access Control (RBAC) system for managing user access to organizations.

### Account Creation & Authentication Flow

**Authentication Middleware** (`internal/middlewares/auth/account_token.go`):

The middleware automatically creates accounts during first login and distinguishes between two flows:

1. **Self-Signup Flow** (No pending invite):
   - Calls `CreateAccountWithAutoOrg()` - creates account + trial org atomically
   - Trial org named with pattern: `${email}-trial`
   - User gets `DefaultEvaluationJourneyWithAutoOrg()` journey tracking
   - Automatically assigned as org admin
   - Skips manual org creation in dashboard

2. **Invite Flow** (Pending invite exists):
   - Calls standard `CreateAccount()` with `NoUserJourneys()`
   - Existing invite mechanisms handle org access
   - No auto org creation

### Permission Architecture (Three-Layer System)

**Layer 1: Accounts** (`internal/app/account.go`)
- Individual users or service accounts
- Types: `Auth0`, `Service`, `Canary`, `Integration`
- `AfterQuery` hook aggregates permissions from all roles

**Layer 2: Roles** (`internal/app/role.go`)
- Permission containers with specific purposes
- Standard types: `OrgAdmin`, `Installer`, `Runner`
- Created per organization via `authzClient.CreateOrgRoles()`

**Layer 3: Policies** (`internal/app/policy.go`)
- Actual permission sets attached to roles
- Stored in PostgreSQL HSTORE format
- Define granular permissions for resources

### Org Role Creation Process

**AuthZ Client Operations** (`internal/pkg/authz/`):

1. **Create Org Roles** (`create_org_roles.go`):
   ```go
   authzClient.CreateOrgRoles(ctx, orgID)
   ```
   - Creates `OrgAdmin`, `Installer`, `Runner` roles
   - Each role gets associated policies with permissions
   - Requires account context for audit trail

2. **Assign User to Org** (`add_account_role.go`):
   ```go
   authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, orgID, accountID)
   ```
   - Creates `AccountRole` junction table entry
   - Links account to specific org role
   - Uses conflict resolution (`DoNothing`) for safety

### Critical Context Requirements

**CreatedByID Audit Pattern**:
All major entities (Org, Role, Policy, AccountRole) have `CreatedByID` fields populated by `BeforeCreate` hooks:

```go
// REQUIRED before any authz operations
ctx = cctx.SetAccountContext(ctx, account)

// Now hooks can find account ID
authzClient.CreateOrgRoles(ctx, orgID)        // ✅ Works
authzClient.AddAccountOrgRole(ctx, ..., ...)  // ✅ Works
```

**Common Issue**: Calling authz methods without account context results in:
- `CreatedByID` fields being empty/null
- Potential constraint violations
- Role creation failures

### Account Permission Resolution

**Runtime Permission Aggregation** (`internal/app/account.go:57-85`):

The `AfterQuery` hook automatically:
1. Aggregates permissions from all account roles
2. Builds `OrgIDs` array of accessible organizations
3. Creates unified `AllPermissions` set for authorization
4. Populates `Orgs` array with accessible org objects

### Auto-Org Implementation Details

**Key Files & Methods**:

- `internal/pkg/account/create.go`:
  - `CreateAccountWithAutoOrg()` - Atomic account + org creation
  - `DefaultEvaluationJourneyWithAutoOrg()` - Journey with org step completed
  - Uses database transactions for consistency

- `internal/middlewares/auth/account_token.go:85-104`:
  - Self-signup detection via pending invite check
  - Context setup for authz operations
  - Error handling with detailed messages

- `internal/app/user_journey.go`:
  - Journey step definitions and JSONB storage
  - Integration with dashboard modal system

### Transaction Safety & Error Handling

**Database Transactions**:
```go
tx := m.db.WithContext(ctx).Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// 1. Create account
// 2. Set account context
// 3. Create org
// 4. Update notifications config
tx.Commit()

// Then outside transaction:
// 5. Create org roles (with context)
// 6. Assign user as admin
```

**Error Patterns**:
- Account creation errors: Return immediately with transaction rollback
- Org creation errors: Return immediately with transaction rollback
- Role creation errors: Return detailed error (org exists but no access)
- Context missing errors: Usually silent failures in hooks

### User Journey Integration

**Updated Journey Steps**:
1. `account_created` - Account signup (auto-completed)
2. `org_created` - Trial org creation (auto-completed for self-signup)
3. `app_created` - App sync via CLI (user action required)
4. `install_created` - Install creation (user action required)

**Cross-Service Updates**:
- Journey updates via `accountsHelpers.UpdateUserJourneyStep*()`
- Dependency injection pattern: helpers accessed in services
- Non-blocking: Journey failures logged but don't break operations

### Testing & Debugging

**Common Issues**:
1. **Missing Context**: Roles created without `CreatedByID` - check context setup
2. **Transaction Errors**: Org exists but user has no access - check role assignment
3. **Journey Not Updated**: User sees org creation modal - check helper calls
4. **Permission Errors**: User can't access org - check `AccountRole` records

**Debug Queries**:
```sql
-- Check account roles
SELECT ar.*, r.role_type FROM account_roles ar
JOIN roles r ON ar.role_id = r.id
WHERE ar.account_id = 'account_id';

-- Check org roles exist
SELECT * FROM roles WHERE org_id = 'org_id';

-- Check user journey state
SELECT user_journeys FROM accounts WHERE id = 'account_id';
```

This RBAC system enables Nuon's multi-tenant architecture while providing secure, auditable, and flexible permission management.

## Global Endpoint Configuration & Debugging

The ctl-api uses a global middleware system to mark endpoints as "global" (not organization-scoped) while still requiring authentication.

### Global Middleware Pattern

**Location**: `internal/middlewares/global/global.go`

**Critical Rule**: The global endpoint list must use **exact route patterns** matching the actual route registration:

```go
// Route registration
api.POST("/v1/account/user-journeys/:journey_name/complete", s.CompleteUserJourney)

// Global middleware list (MUST match exactly)
{"POST", "/v1/account/user-journeys/:journey_name/complete"}: {},
```

### Common Configuration Issues

**Problem**: Route pattern mismatch between global middleware and route registration
```go
// ❌ WRONG - Literal path in global middleware
{"POST", "/v1/account/user-journeys/evaluation/complete"}: {},

// ✅ CORRECT - Parameterized path matching route registration
{"POST", "/v1/account/user-journeys/:journey_name/complete"}: {},
```

**Symptoms of Mismatch**:
- API endpoints return 401/403 authentication errors
- Requests fail organization context validation
- Frontend experiences "silent failures" with no visible errors
- Global endpoints incorrectly require `orgId` parameter

### Debugging Global Endpoint Issues

**Detection**: Check if endpoint is properly marked as global:
1. Verify route pattern in global middleware matches route registration exactly
2. Check middleware logs for "marking request as global" debug messages
3. Test API endpoint with/without `X-Nuon-Org-ID` header

**Resolution Pattern**:
```go
// 1. Find route registration pattern
api.METHOD("/v1/path/:param/endpoint", handler)

// 2. Use identical pattern in global middleware
{"METHOD", "/v1/path/:param/endpoint"}: {},
```

### Authentication Context Requirements

**Global Endpoints**:
- Still require valid JWT authentication via `Authorization: Bearer <token>`
- Do NOT require `X-Nuon-Org-ID` header
- Marked as global via middleware for context validation bypass

**Organization-Scoped Endpoints**:
- Require both JWT authentication AND `X-Nuon-Org-ID` header
- Subject to organization access validation
- Account must have appropriate roles for the organization

This global endpoint system enables account-level operations while maintaining the multi-tenant security model.
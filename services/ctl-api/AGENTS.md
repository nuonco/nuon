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
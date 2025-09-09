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
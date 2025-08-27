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

## Getting Started

1. **Prerequisites**: Go 1.24+, Node.js, Docker, Terraform, kubectl
2. **Authentication**: Set up cloud credentials (AWS/Azure)  
3. **Local Development**: Check service-specific README files in `/services/` and `/bins/`
4. **Documentation**: Visit the `/docs/` directory or internal wiki

## Notes for Claude

- This is a complex enterprise platform with many interconnected services
- The main business logic is in Go, with TypeScript for UIs
- Heavy use of Kubernetes, Terraform, and cloud-native technologies
- The `/bins/nuonctl/scripts/` directory contains many operational scripts
- Infrastructure code is in `/infra/` with Terraform modules
- Example applications and templates are in `/seed/`

## Project Status

Main branch: `main`
Repository is clean with recent commits related to CLI improvements and authentication.
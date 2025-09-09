# Dashboard UI Service  

The **Dashboard UI** is the primary web application frontend for the Nuon platform, providing a comprehensive interface for managing applications, deployments, and infrastructure.

## Service Overview

This is the main user-facing web application built with Next.js and React. It provides a complete dashboard experience for developers and operators to manage their BYOC (Bring Your Own Cloud) deployments through the Nuon platform.

## Architecture

- **Framework**: Next.js 15+ with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS with custom design system
- **State Management**: React hooks with API data fetching
- **Authentication**: Auth0 integration via `@auth0/nextjs-auth0`
- **Testing**: Vitest with React Testing Library
- **Build Tool**: Turbo for development speed

## Relationship to Other Services

- **Primary API**: Consumes `ctl-api` service for all backend operations
- **Authentication**: Integrates with Auth0 for user authentication
- **Monitoring**: DataDog RUM and logging integration
- **Analytics**: Segment analytics for user behavior tracking
- **Infrastructure**: Deployed to Kubernetes via Helm charts

## Project Structure

### Core Files
- `next.config.mjs` - Next.js configuration
- `tailwind.config.ts` - Tailwind CSS customization
- `package.json` - Dependencies and scripts
- `Dockerfile` - Container build configuration

### Key Directories

#### `/src/app/` - Next.js App Router
Modern Next.js file-based routing structure:
- `[org-id]/` - Organization-scoped routes
  - `apps/[app-id]/` - Application management pages
  - `installs/[install-id]/` - Installation management and monitoring
  - `releases/` - Release management interface
  - `runner/` - Runner status and management
  - `team/` - Organization team management
- `api/` - API route handlers and proxy endpoints
- `stratus/` - New UI framework implementation

#### `/src/components/` - React Components
Comprehensive component library organized by domain:
- `Apps/` - Application-specific components
- `Components/` - Build and component management
- `Installs/` - Installation management UI
- `InstallComponents/` - Deployment component management
- `InstallWorkflows/` - Workflow and approval interfaces
- `Runners/` - Runner monitoring and management
- `Orgs/` - Organization administration
- `LogStream/` - Real-time log viewing
- Common UI components: `Button`, `Modal`, `DataTable`, etc.

#### `/src/stratus/` - New Design System
Modern React component library:
- `components/` - Reusable UI components
- `context/` - React context providers
- `actions/` - Server actions for data mutations

#### `/src/lib/` - Business Logic
API client libraries and utilities:
- `apps.ts` - Application management logic
- `installs.ts` - Installation operations
- `components.ts` - Component build management
- `orgs.ts` - Organization operations
- `runners.ts` - Runner communication

#### `/src/utils/` - Utility Functions
- `auth.ts` - Authentication helpers
- `query-data.ts` - API data fetching
- `mutate-data.ts` - Data mutation helpers
- `time-utils.ts` - Date/time formatting
- `datadog-*.tsx` - Monitoring integration

#### `/src/types/` - TypeScript Definitions
- `nuon-oapi-v3.d.ts` - Auto-generated API types from OpenAPI spec
- `ctl-api.types.ts` - Custom API type definitions
- `dashboard.types.ts` - Dashboard-specific types

### Infrastructure

#### `/infra/` - Terraform Configuration
- `service.tf` - ECS/EKS service definition
- `certificate.tf` - SSL certificate management
- Infrastructure deployment configuration

#### `/k8s/` - Kubernetes Deployment
Helm chart for Kubernetes deployment:
- `templates/` - Kubernetes resource templates
- `values.yaml` - Configuration values

## Key Features

### Multi-Organization Support
- Organization switching and context management
- Role-based access control and permissions
- Team member management and invitations

### Application Management
- App creation and configuration
- Component dependency management
- Build history and status tracking
- Configuration templating and validation

### Installation Management
- Infrastructure provisioning workflows
- Component deployment orchestration
- Real-time status monitoring
- Approval workflows for sensitive operations

### Workflow Management
- Visual workflow execution tracking
- Step-by-step approval processes
- Terraform plan reviewing and approval
- Deployment rollback capabilities

### Monitoring & Observability  
- Real-time log streaming
- Runner health monitoring
- Deployment status tracking
- Infrastructure state visualization

### Developer Experience
- Code editor integration (Monaco)
- Terraform/YAML syntax highlighting
- Configuration diff viewing
- Template rendering and validation

## Development

### Setup
```bash
cd services/dashboard-ui
npm install
npm run dev
```

### Key Scripts
- `npm run dev` - Development server with Turbo
- `npm run build` - Production build
- `npm run generate-api-types` - Generate types from OpenAPI spec
- `npm run test` - Run tests with Vitest
- `npm run lint` - ESLint validation

### API Integration
The dashboard automatically generates TypeScript types from the ctl-api OpenAPI specification, ensuring type safety across the entire application.

### Testing
- Unit tests with Vitest
- Component tests with React Testing Library
- API mocking with MSW (Mock Service Worker)

## Configuration

### Environment Variables
- `NUON_API_URL` - Backend API endpoint
- Auth0 configuration for authentication
- DataDog keys for monitoring
- Feature flags for experimental features

### Build Configuration
- Next.js App Router configuration
- Tailwind CSS with custom design tokens
- TypeScript strict mode enabled
- Bundle analysis and optimization

## Deployment

The dashboard is containerized and deployed to Kubernetes:
- Docker multi-stage build for optimization
- Helm chart deployment with environment-specific values
- SSL termination and custom domain support
- Auto-scaling based on traffic

This service provides the primary user interface for the entire Nuon platform, enabling users to manage complex cloud deployments through an intuitive web interface.
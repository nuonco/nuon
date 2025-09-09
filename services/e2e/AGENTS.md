# E2E Service

The **E2E (End-to-End)** service provides comprehensive end-to-end testing capabilities for the Nuon platform, including API testing, infrastructure validation, and full deployment workflow testing.

## Service Overview

This is a sophisticated testing service built in Go that performs comprehensive end-to-end testing of the Nuon platform. It includes both API and worker components to test various aspects of the platform from infrastructure provisioning to application deployment.

## Architecture

- **Language**: Go
- **Components**: API server and worker processes
- **Testing Framework**: Custom testing framework with Go
- **Infrastructure**: Uses real cloud resources for testing
- **Orchestration**: Temporal workflow integration
- **Containerization**: Docker with Kubernetes deployment

## Relationship to Other Services

- **Platform Testing**: Tests all major Nuon services and workflows
- **Infrastructure Validation**: Validates `ctl-api` and infrastructure components
- **CI/CD Integration**: Used in continuous integration pipelines
- **Quality Assurance**: Ensures platform reliability and functionality
- **Regression Testing**: Prevents regressions in new releases

## Project Structure

### Core Components

#### `/api/` - API Server
- `main.go` - API server entry point
- `discover.go` - Test discovery and execution
- `internal/` - Internal API logic

#### `/worker/` - Background Workers
- `main.go` - Worker process entry point
- Background test execution
- Resource cleanup and management

#### `/nuon/` - Nuon Platform Integration
Terraform configurations for testing platform components:
- `app.tf` - Application creation and management
- `component_*.tf` - Various component types testing
- `installer.tf` - Installer functionality testing
- `installs.tf` - Installation process testing
- Lifecycle scripts for setup and teardown

#### `/infra/` - Infrastructure Setup
- `ecr.tf` - Container registry setup
- `bucket.tf` - S3 storage for testing
- `data.tf` - Data sources and variables
- Infrastructure for the testing environment

### Infrastructure Components

#### `/chart/` - Kubernetes Deployment
Helm chart for deploying the E2E service:
- `templates/` - Kubernetes resource templates
- API and worker deployment configurations
- Service accounts and RBAC
- Auto-scaling configuration

#### `/infra-empty/` - Minimal Infrastructure
Empty infrastructure setup for basic testing scenarios

## Key Features

### Comprehensive Testing
- Full platform workflow testing
- Infrastructure provisioning validation
- Component build and deployment testing
- Integration testing across services

### Real Infrastructure Testing
- Uses actual cloud resources (AWS, Azure)
- Tests real deployment scenarios
- Validates infrastructure as code
- End-to-end workflow validation

### Test Discovery & Execution
- Automatic test discovery
- Parallel test execution
- Test result reporting
- Cleanup and resource management

### Workflow Integration
- Temporal workflow testing
- Long-running test scenarios
- Complex orchestration validation
- Failure recovery testing

## Development

### Setup
```bash
cd services/e2e
go mod download
```

### Running Tests
- API Server: `go run api/main.go`
- Worker: `go run worker/main.go`
- Specific tests: Configure via Nuon platform

### Test Configuration
Tests are configured through:
- Environment variables
- Nuon configuration files
- Terraform variables
- Test-specific parameters

## Testing Scenarios

### Platform Integration Tests
- User and organization management
- Application lifecycle testing
- Component builds and releases
- Installation workflows

### Infrastructure Tests
- Cloud resource provisioning
- Terraform execution validation
- Kubernetes deployment testing
- Network and security validation

### API Testing
- REST API endpoint validation
- Authentication and authorization
- Data consistency checks
- Performance and load testing

### Workflow Testing
- Complex deployment scenarios
- Multi-step approval processes
- Error handling and recovery
- Resource cleanup validation

## Deployment

### Kubernetes Deployment
- Helm chart with configurable values
- Separate API and worker deployments
- Auto-scaling based on test load
- Resource limits and monitoring

### Infrastructure Requirements
- Access to cloud providers (AWS, Azure)
- ECR for container images
- S3 for test artifacts
- Proper IAM roles and permissions

## Configuration

### Environment Variables
- Cloud provider credentials
- Nuon API endpoints
- Test configuration parameters
- Resource limits and timeouts

### Test Management
- Test suite configuration
- Parallel execution settings
- Resource cleanup policies
- Reporting and alerting setup

## Technologies Used

### Core Technologies
- **Go**: Primary language for all components
- **Temporal**: Workflow orchestration
- **Terraform**: Infrastructure testing
- **Kubernetes**: Container orchestration

### Testing Framework
- Custom Go testing framework
- Infrastructure validation tools
- API testing utilities
- Resource management helpers

### Cloud Integration
- AWS SDK for cloud resource management
- Azure SDK for multi-cloud testing
- Kubernetes client libraries
- Container registry integration

This service ensures the reliability and quality of the entire Nuon platform by providing comprehensive end-to-end testing capabilities that validate everything from basic API functionality to complex multi-cloud deployment scenarios.
# Workers - Canary Service

The **Workers Canary** service provides automated canary testing and validation for the Nuon platform, creating and managing test environments to ensure platform stability and functionality.

## Service Overview

This is a Go-based worker service that creates and manages canary test environments for the Nuon platform. It provisions test organizations, runs automated tests, and validates that new releases work correctly before they reach production users.

## Architecture

- **Language**: Go
- **Framework**: Temporal workflows for orchestration
- **Testing**: Automated canary environment management
- **Infrastructure**: Creates real cloud resources for testing
- **Lifecycle Management**: Full provisioning and cleanup automation

## Relationship to Other Services

- **Quality Assurance**: Validates `ctl-api` and platform functionality
- **Release Management**: Part of the deployment pipeline
- **Infrastructure Testing**: Uses `workers-executors` for resource management
- **Platform Validation**: Tests all major platform workflows
- **Automated Testing**: Integrates with CI/CD pipelines

## Project Structure

### Core Files
- `main.go` - Service entry point
- `service.yml` - Service configuration
- `install-cli.sh` - CLI installation script

### Key Directories

#### `/cmd/` - Command Structure
- `all.go` - All commands aggregation
- `root.go` - Root command and CLI setup

#### `/internal/` - Core Logic

##### `/internal/activities/` - Temporal Activities
- `activities.go` - Activity definitions
- `create_canary_user.go` - User creation for testing
- `create_org.go` - Organization provisioning
- `create_vcs_connection.go` - VCS integration testing
- `delete_org.go` - Cleanup operations
- `exec_test_script.go` - Test script execution
- `list_tests.go` - Test discovery
- `poll_active_installs.go` - Installation monitoring
- `terraform.go` - Infrastructure operations

##### `/internal/workflows/` - Temporal Workflows
- `provision.go` - Environment provisioning workflow
- `deprovision.go` - Cleanup and teardown workflow
- `exec.go` - Test execution workflow
- `tests.go` - Test orchestration
- `notifications.go` - Alert and notification handling

#### `/terraform/` - Infrastructure Definitions
Test infrastructure configurations:
- `eks.tf` - Amazon EKS cluster testing
- `ecs.tf` - Amazon ECS service testing
- `aks.tf` - Azure Kubernetes Service testing
- `main.tf` - Core infrastructure setup
- `e2e/` - End-to-end test scenarios

#### `/tests/` - Test Configurations
Comprehensive test suites:
- `001-orgs` - Organization management tests
- `002-apps` - Application lifecycle tests
- `003-config-file-default` - Default configuration testing
- `004-config-file-sources` - Source configuration validation
- `005-config-file-components` - Component configuration testing
- `006-config-file-min` - Minimal configuration tests
- `007-installs` - Installation process testing
- `008-components` - Component functionality tests
- `009-builds` - Build process validation
- `010-config-file-actions` - Action configuration testing
- `011-releases` - Release management testing
- `012-installs` - Advanced installation scenarios

#### Infrastructure & Deployment

##### `/infra/` - Service Infrastructure
- `service.tf` - ECS/EKS service definition
- `bucket.tf` - S3 storage for artifacts
- `install_access_role.tf` - IAM roles for testing

##### `/k8s/` - Kubernetes Deployment
Helm chart for Kubernetes deployment:
- `templates/` - Resource templates
- `values.yaml` - Configuration values

## Key Features

### Automated Canary Testing
- Automatic test environment creation
- Full platform workflow validation
- Multi-cloud infrastructure testing
- Comprehensive cleanup after testing

### Temporal Workflow Orchestration
- Complex multi-step test scenarios
- Parallel test execution
- Error handling and retry logic
- Resource cleanup on failure

### Infrastructure Validation
- Real cloud resource provisioning
- EKS, ECS, and AKS testing
- Network and security validation
- Performance and reliability testing

### Test Management
- Configurable test suites
- Automatic test discovery
- Result reporting and alerting
- Integration with monitoring systems

## Development

### Setup
```bash
cd services/workers-canary
go mod download
```

### Running Locally
```bash
go run main.go
```

### Configuration
- Environment variables for cloud credentials
- Temporal cluster configuration
- Test parameters and timeouts
- Notification and alerting setup

## Test Scenarios

### Platform Core Tests
- Organization and user management
- Application creation and configuration
- Component builds and releases
- Installation and deployment workflows

### Infrastructure Tests
- Cloud resource provisioning
- Kubernetes cluster management
- Container deployment validation
- Network connectivity testing

### Configuration Tests
- TOML configuration file validation
- Default configuration testing
- Component dependency resolution
- Action and workflow configuration

## Deployment

### Kubernetes Deployment
- Helm chart with auto-scaling
- Resource limits and monitoring
- Health checks and readiness probes
- Secret and configuration management

### Infrastructure Requirements
- Multi-cloud access (AWS, Azure)
- Temporal cluster connectivity
- Container registry access
- Monitoring and alerting integration

## Configuration Files

### Test Configurations
Located in `/tests/configs/`:
- `nuon.default.toml` - Default configuration testing
- `nuon.components.toml` - Component-specific testing
- `nuon.actions.toml` - Action workflow testing
- `nuon.sources.toml` - Source integration testing
- `nuon.min.toml` - Minimal configuration testing

## Technologies Used

### Core Technologies
- **Go**: Primary service language
- **Temporal**: Workflow orchestration
- **Terraform**: Infrastructure provisioning
- **Kubernetes**: Container orchestration

### Cloud Integrations
- **AWS**: EKS, ECS, S3, IAM
- **Azure**: AKS, storage, networking
- **Docker**: Container management
- **Helm**: Kubernetes deployment

This service ensures platform reliability by continuously testing new releases in realistic environments, catching issues before they reach production users and maintaining high quality standards for the Nuon platform.
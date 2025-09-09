# Runner Binary

The **Runner** is a critical binary that executes deployment operations in customer infrastructure. Unlike other binaries, it runs as an executable in both Kubernetes containers and cloud VMs (AWS, Azure, with more platforms planned), providing secure execution of deployments in customer environments.

## Binary Overview

The Runner is the execution engine that runs within customer infrastructure to perform deployments, manage infrastructure state, and execute workflows. It operates in customer-controlled environments while maintaining secure communication with the Nuon control plane.

## Architecture

- **Language**: Go
- **Deployment**: Kubernetes containers and cloud VMs (AWS, Azure, future: GCP, etc.)
- **Execution Model**: Job-based execution with lifecycle management
- **Security**: Operates within customer security boundaries
- **Communication**: Secure API communication with Nuon control plane
- **State Management**: Manages Terraform state, Helm releases, and deployments

## Relationship to Other Services

- **API Communication**: Communicates with `ctl-api` Runner API endpoints
- **Job Execution**: Executes jobs queued by the control plane
- **State Reporting**: Reports execution status and results back to platform
- **Customer Infrastructure**: Operates within customer cloud accounts
- **Security Bridge**: Provides secure execution in customer environments

## Project Structure

### Core Files
- `main.go` - Runner binary entry point
- `Dockerfile` - Container image for Kubernetes deployment
- `build-config.yaml` - Build configuration
- `install.sh` - Installation script for VM deployment
- `generate.sh` - Code generation script
- `service.yml` - Service configuration

### Key Directories

#### `/cmd/` - Command Structure
- `root.go` - Root command and global configuration
- `cli.go` - CLI command definitions
- `run.go` - Main execution loop
- `run_local.go` - Local development execution
- `mng.go` - Management operations
- `install.go` - Installation and setup
- `build.go` - Build operations
- `version.go` - Version information

#### `/internal/` - Core Logic

##### Job Execution (`/jobs/`)
The heart of the runner - job execution framework:

###### Action Workflows (`/actions/`)
- `actions.go` - Action workflow execution
- `workflow/` - Workflow step execution
  - `build.go` - Build step execution
  - `exec.go` - Command execution
  - `fetch.go` - Resource fetching
  - `init.go` - Initialization
  - `outputs.go` - Output management
  - `state.go` - State management
  - `validate.go` - Validation

###### Deployment Operations (`/deploy/`)
Multi-type deployment handlers:
- `helm/` - Helm chart deployments
  - `client.go` - Helm client integration
  - `diff.go` - Deployment diffs
  - `operation_*.go` - Install/upgrade/uninstall
  - `outputs.go` - Helm deployment outputs
- `terraform/` - Terraform deployments
  - `exec.go` - Terraform execution
  - `workspace.go` - Workspace management  
  - `outputs.go` - Terraform outputs
  - `graceful_shutdown.go` - Clean shutdown
- `kubernetes_manifest/` - Raw Kubernetes manifests
  - `client.go` - Kubernetes API client
  - `exec.go` - Manifest application
  - `diff.go` - Resource diffing
- `job/` - Kubernetes Job deployments
  - `deploy.go` - Job deployment
  - `monitor.go` - Job monitoring

###### Sandbox Operations (`/sandbox/`)
- `terraform/` - Sandbox Terraform operations
- `sync_secrets/` - Secret synchronization

###### Management Operations (`/management/`)
- `update/` - Runner updates
- `shutdown/` - Graceful shutdown
- `vm_shutdown/` - VM shutdown operations

###### Health & Monitoring (`/healthcheck/`)
- `check/` - Health check execution
- Regular health reporting to control plane

##### Core Packages (`/pkg/`)

###### API Integration (`/api/`)
- `api.go` - Runner API client for control plane communication

###### Development Support (`/dev/`)
- `dev.go` - Local development mode
- `runner.go` - Development runner setup
- `monitor.go` - Development monitoring

###### Job Processing (`/jobloop/`)
- `jobloop.go` - Main job processing loop
- `exec_job.go` - Job execution coordination
- `job_handler.go` - Job handling logic
- `monitor_job.go` - Job monitoring
- `worker.go` - Worker management

###### Infrastructure Management
- `k8s/` - Kubernetes integration
- `workspace/` - Workspace and directory management
- `git/` - Git repository operations
- `oci/` - OCI/container operations

###### Observability
- `log/` - Logging infrastructure
- `metrics/` - Metrics collection
- `exporter/` - Telemetry export (OTEL)
- `heartbeater/` - Heartbeat management

###### Resource Management
- `outputs/` - Output processing (Terraform, Helm, etc.)
- `registry/` - Container registry operations
- `settings/` - Runner settings management
- `plan/` - Deployment plan processing

#### `/bundle/` - Deployment Bundles
Pre-configured deployment templates:

##### Kubernetes Deployment (`/helm/`)
- `Chart.yaml` - Helm chart definition
- `templates/` - Kubernetes resource templates
  - `deployment.tpl` - Runner deployment
  - `config_map.tpl` - Configuration
  - `rbac.tpl` - Role-based access control
  - `node_pool.tpl` - Node pool configuration
- `values.yaml` - Default configuration values

##### ECS Deployment (`/terraform-ecs/`)
- Complete Terraform configuration for AWS ECS deployment
- Service definitions and networking
- IAM roles and security configuration

## Key Features

### Multi-Platform Execution
- **Kubernetes**: Runs as containerized workloads
- **AWS VMs**: Native EC2 instance execution
- **Azure VMs**: Azure Virtual Machine execution
- **Future Platforms**: Designed for GCP and other cloud providers

### Job Execution Engine
- **Job Processing**: Processes jobs from the control plane queue
- **Lifecycle Management**: Complete job lifecycle from fetch to completion
- **State Management**: Manages deployment state and outputs
- **Error Handling**: Robust error handling and recovery

### Deployment Capabilities
- **Terraform**: Full Terraform workflow execution
- **Helm**: Kubernetes Helm chart deployments
- **Kubernetes Manifests**: Raw Kubernetes resource management
- **Container Jobs**: Kubernetes job execution
- **Custom Actions**: Extensible action workflow system

### Security & Isolation
- **Customer Boundary**: Operates within customer security perimeters
- **Secure Communication**: Encrypted communication with control plane
- **Credential Management**: Secure handling of cloud credentials
- **Audit Logging**: Comprehensive audit trail

### Observability
- **Metrics**: Runtime and deployment metrics
- **Logging**: Structured logging with OTEL integration
- **Tracing**: Distributed tracing for operations
- **Health Monitoring**: Continuous health reporting

## Deployment Modes

### Kubernetes Deployment
- **Helm Chart**: Pre-configured Helm chart in `/bundle/helm/`
- **Container Image**: Docker image with all dependencies
- **Resource Management**: CPU/memory limits and auto-scaling
- **RBAC**: Proper Kubernetes permissions

### VM Deployment (AWS/Azure)
- **Installation Script**: Automated installation via `install.sh`
- **Service Management**: Systemd service configuration
- **Auto-updates**: Automatic runner updates
- **Monitoring**: VM health and status monitoring

### Local Development
- **Dev Mode**: `nuonctl` integration for local testing
- **Hot Reload**: Development workflow with file watching
- **Mock Execution**: Local job execution for testing

## Development

### Setup
```bash
cd bins/runner
go build -o runner .
```

### Local Testing
```bash
./runner run-local  # Local development mode
./runner --help     # Available commands
```

### Building Container
```bash
docker build -t nuon-runner .
```

## Configuration

### Environment Variables
- Runner identification and registration
- Control plane API endpoints  
- Cloud provider credentials
- Execution environment settings

### Job Configuration
- Job-specific parameters and inputs
- Resource limits and timeouts
- Output destinations and formats
- Error handling policies

## Execution Flow

### Job Lifecycle
1. **Registration**: Runner registers with control plane
2. **Job Polling**: Continuously polls for available jobs
3. **Job Fetch**: Downloads job specification and resources
4. **Initialization**: Sets up execution environment
5. **Execution**: Runs deployment operations
6. **Output Collection**: Collects and processes outputs
7. **Status Reporting**: Reports progress and completion
8. **Cleanup**: Cleans up temporary resources

### Deployment Process
1. **Plan Generation**: Creates deployment plans
2. **Validation**: Validates configurations and permissions
3. **Resource Provisioning**: Creates/updates cloud resources
4. **Application Deployment**: Deploys applications and services
5. **Verification**: Verifies deployment success
6. **State Management**: Updates deployment state

## Technologies Used

### Core Technologies
- **Go**: Primary language with extensive library ecosystem
- **Terraform**: Infrastructure as Code execution
- **Helm**: Kubernetes package management
- **Kubernetes**: Container orchestration client libraries

### Cloud Integration
- **AWS SDK**: Amazon Web Services integration
- **Azure SDK**: Microsoft Azure integration
- **Docker**: Container runtime and registry
- **OCI**: Open Container Initiative standards

### Observability
- **OpenTelemetry**: Metrics, logging, and tracing
- **Prometheus**: Metrics collection
- **Structured Logging**: JSON-based logging
- **Distributed Tracing**: End-to-end request tracing

The Runner is the critical execution component that enables Nuon to securely deploy and manage applications within customer infrastructure while maintaining the security and isolation requirements of enterprise customers.
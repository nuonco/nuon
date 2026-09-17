# Runner Binary

The **Runner** is a critical binary that executes deployment operations in customer infrastructure. Unlike other
binaries, it runs as an executable in both Kubernetes containers and cloud VMs (AWS, Azure, GCP), providing secure
execution of deployments in customer environments.

## Binary Overview

The Runner is the execution engine that runs within customer infrastructure to perform deployments, manage
infrastructure state, and execute workflows. It operates in customer-controlled environments while maintaining secure
communication with the Nuon control plane.

## Architecture

- **Language**: Go
- **Deployment**: Kubernetes containers and cloud VMs (AWS, Azure, GCP)
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

## Capabilities

Runs in Kubernetes and cloud VMs (AWS, Azure, GCP). Executes Terraform, Helm, Kubernetes manifests, container jobs, and
action workflows. Reports status to the ctl-api runner API; emits OTEL metrics/logs.

## Deployment Modes

- **Kubernetes**: Helm chart under `bundle/helm/`
- **VM**: `install.sh` + systemd (`install` / `mng` modes)
- **Local**: `./runner run-local`

## Development

```bash
cd bins/runner
go build -o runner .
./runner --help
./runner run-local
docker build -t nuon-runner .
```

## Execution Flow

1. Register with control plane → poll for jobs
2. Fetch job spec and resources → initialize workspace
3. Execute (plan/apply/action) → collect outputs → report status → cleanup


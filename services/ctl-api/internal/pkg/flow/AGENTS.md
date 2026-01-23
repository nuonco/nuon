# Flow Package

The flow package orchestrates workflow step execution in Temporal workflows. It handles step lifecycle, approval flows, and status updates.

## Key Files

- `conductor.go` - Generic workflow conductor that drives step execution
- `execute_flow_step.go` - Core step execution logic including approval handling
- `execute_workflow_step.go` - Legacy step execution (being migrated)
- `approval.go` - Approval signal handling and response processing
- `rerun_flow.go` - Re-execution logic for failed/retried steps

## Step Target Resolution

Workflow steps use polymorphic relationships to track their execution targets:

### StepTargetType Values
- `install_sandbox_runs` → Sandbox terraform operations
- `install_deploys` → Component deploy operations
- `install_action_workflow_runs` → Action workflow executions
- `install_cloudformation_stack` → CloudFormation stack operations

### Data Chain for Component Deploys

```
StepTargetID (install deploy ID)
    └── InstallDeploy
            ├── ComponentBuild
            └── InstallComponent
                    ├── Component (has Type: terraform_module, helm_chart, etc.)
                    └── Install (has AppConfigID)
```

**Activity**: `GetDeploy` in `installs/worker/activities/get_deploy.go` preloads this full chain.

### Data Chain for Sandbox Runs

```
StepTargetID (sandbox run ID)
    └── InstallSandboxRun
            ├── InstallSandbox
            └── Install (has AppConfigID)
```

**Activities**: 
- `GetSandboxRun` in `installs/worker/activities/get_sandbox_run.go`
- `GetInstallForSandbox` to get the Install from sandbox ID

## Approval Step Flow

For steps with `ExecutionType == WorkflowStepExecutionTypeApproval`:

1. Step executes (creates plan via runner job)
2. `StepTargetID` is set to the created InstallDeploy or InstallSandboxRun
3. `CheckNoopPlan` activity evaluates if plan has changes
4. If noop → auto-skip step and next apply step
5. If has changes → wait for approval signal
6. On approval → proceed to apply step

## Policy Configuration Access

Policies are loaded through the AppConfig chain:

```
Install.AppConfigID
    └── AppConfig
            └── PoliciesConfig (AppPoliciesConfig)
                    └── Policies []AppPolicyConfig
                            ├── Type (terraform_module, helm_chart, kubernetes_cluster, etc.)
                            ├── Engine (opa, kyverno)
                            ├── Contents (policy document)
                            ├── Components []string (component names, empty = all)
                            └── Sandbox bool
```

**Activity**: `GetAppConfig` in `installs/worker/activities/get_app_config.go` loads full config with policies.

## Plan Contents Access

Plan contents (for noop checking and policy evaluation) come from runner job execution results:

```
StepTargetID
    └── RunnerJob (polymorphic Owner relation)
            └── RunnerJobExecution (latest by created_at)
                    └── RunnerJobExecutionResult
                            └── ContentsDisplayGzip (gzip compressed plan JSON)
```

This is implemented in `getApprovalPlan` in `workflows/workflow/activities/check_noop_plan.go`.

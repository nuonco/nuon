package sandbox

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

func (w *Workflows) executeSandboxPlan(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, stepID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	op := app.RunnerJobOperationTypeCreateApplyPlan
	if installRun.RunType == app.SandboxRunTypeDeprovision {
		op = app.RunnerJobOperationTypeCreateTeardownPlan
	}

	runnerJob, err := activities.AwaitCreateSandboxJob(ctx, &activities.CreateSandboxJobRequest{
		InstallID: install.ID,
		RunnerID:  install.RunnerID,
		OwnerType: "install_sandbox_runs",
		OwnerID:   installRun.ID,
		Op:        op,
		Metadata: map[string]string{
			"install_id":       install.ID,
			"sandbox_run_id":   installRun.ID,
			"sandbox_run_type": string(installRun.RunType),
		},
	})
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	runPlan, err := plan.AwaitCreateSandboxRunPlan(ctx, &plan.CreateSandboxRunPlanRequest{
		RunID:      installRun.ID,
		InstallID:  install.ID,
		RootDomain: w.cfg.DNSRootDomain,
		WorkflowID: fmt.Sprintf("%s-create-api-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan request")
		return errors.Wrap(err, "unable to create plan")
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	// Get app config for role selection
	appConfig, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to get app config")
		return fmt.Errorf("unable to get app config: %w", err)
	}

	// Get install stack for auth configuration
	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, install.ID)
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to get install stack")
		return errors.Wrap(err, "unable to get install stack")
	}

	// Build composite plan with auth information
	compositePlan := plantypes.CompositePlan{
		SandboxRunPlan: runPlan,
	}

	roleSelection, operation, err := w.getRoleForSandbox(ctx, l, appConfig, installRun, stack)
	if err != nil {
		return errors.Wrap(err, "unable to evaluate role for sandbox")
	}

	l.Info("selected role for sandbox run",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("operation", string(operation)),
		zap.String("run_type", string(installRun.RunType)),
	)

	// Create auth configuration
	planAuth, err := plan.CreatePlanAuth(
		stack.InstallStackOutputs,
		roleSelection.RoleARN,
		fmt.Sprintf("sandbox-run-%s", installRun.ID),
	)
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create auth config")
		return errors.Wrap(err, "unable to create plan auth")
	}
	compositePlan.Auth = planAuth

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: compositePlan,
	}); err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to save plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	l.Info("queued job and waiting on it to be picked up by runner event loop")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:      runnerJob.ID,
		RunnerID:   install.RunnerID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "job failed")
		return fmt.Errorf("unable to execute job: %w", err)
	}
	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "job failed with status"+string(status))
	}

	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job")
	}

	if _, err := activities.AwaitCreateStepApproval(ctx, &activities.CreateStepApprovalRequest{
		OwnerID:     installRun.ID,
		OwnerType:   "install_sandbox_runs",
		RunnerJobID: job.ID,
		StepID:      stepID,
		Type:        app.TerraformPlanApprovalType,
	}); err != nil {
		return errors.Wrap(err, "unable to create approval")
	}

	return nil
}

func (w *Workflows) getRoleForSandbox(
	ctx workflow.Context,
	l *zap.Logger,
	appConfig *app.AppConfig,
	installRun *app.InstallSandboxRun,
	stack *app.InstallStack,
) (*operationroles.RoleSelection, app.OperationType, error) {
	// Determine operation type based on run type
	var operation app.OperationType
	switch installRun.RunType {
	case app.SandboxRunTypeProvision:
		operation = app.OperationProvision
	case app.SandboxRunTypeReprovision:
		operation = app.OperationReprovision
	case app.SandboxRunTypeDeprovision:
		operation = app.OperationDeprovision
	default:
		operation = app.OperationProvision
	}

	// Determine default role based on operation
	defaultRole := appConfig.PermissionsConfig.ProvisionRole.Name
	if operation == app.OperationDeprovision {
		defaultRole = appConfig.PermissionsConfig.DeprovisionRole.Name
	}

	// Get entity roles from sandbox config (if exists)
	// TODO: Load sandbox config and get OperationRoles from it
	var entityRoles operationroles.EntityOperationRoleMap

	// Select role using operation roles engine
	roleSelection, err := operationroles.SelectRole(
		&operationroles.SelectionContext{
			Operation:     operation,
			PrincipalType: principal.TypeSandbox,
			PrincipalName: "", // Sandboxes don't have names
			RuntimeRole:   installRun.Role,
			EntityRoles:   entityRoles,
			MatrixRules:   appConfig.OperationRoleConfig.Rules,
			DefaultRole:   defaultRole,
			AppConfig:     appConfig,
			StackOutputs:  &stack.InstallStackOutputs,
		})
	if err != nil {
		w.updateRunStatusWithoutStatusSync(
			ctx,
			installRun.ID,
			app.SandboxRunStatusError,
			"unable to select role",
		)
		return nil, "", fmt.Errorf("unable to select role: %w", err)
	}

	return roleSelection, operation, nil
}

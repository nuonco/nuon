package worker

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @execution-timeout 30m
func (w *Workflows) SyncSecrets(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: sreq.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		StepID: sreq.WorkflowStepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	if sreq.SandboxMode {
		l.Debug("skipping sync secrets in sandbox mode")
		return nil
	}

	runtimeRole := sreq.SyncSecretsSubSignal.Role

	l.Info("creating plan")
	syncPlan, err := plan.AwaitCreateSyncSecretsPlan(ctx, &plan.CreateSyncSecretsPlanRequest{
		InstallID:   install.ID,
		WorkflowID:  fmt.Sprintf("%s-create-sync-secrets-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
		RuntimeRole: runtimeRole,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create plan")
	}

	if len(syncPlan.KubernetesSecrets) < 1 {
		l.Debug("no secrets to sync")
		return nil
	}

	l.Info("fetching app config for role selection")
	appConfig, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	l.Info("fetching install stack for role selection")
	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, install.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install stack")
	}

	l.Info("fetching install state for role selection")
	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get install state")
	}

	roleSelection, err := w.getRoleForSecretSync(l, appConfig, runtimeRole, stack, installState)
	if err != nil {
		return errors.Wrap(err, "unable to get role for secret sync")
	}

	planAuth, err := plan.CreatePlanAuth(stack.InstallStackOutputs, roleSelection.RoleARN, fmt.Sprintf("secrets-sync-%s", install.ID))
	if err != nil {
		return errors.Wrap(err, "unable to create plan auth")
	}

	// create the job
	l.Info("creating sync secrets job")
	runnerJob, err := activities.AwaitCreateSyncSecretsJob(ctx, &activities.CreateSyncSecretsJobRequest{
		RunnerID:  install.RunnerID,
		InstallID: install.ID,
		OwnerType: "install_workflow_steps",
		OwnerID:   sreq.WorkflowStepID,
		Op:        app.RunnerJobOperationTypeExec,
		Metadata: map[string]string{
			"install_id": install.ID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	planJSON, err := json.Marshal(syncPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SyncSecretsPlan: syncPlan,
			Auth:            planAuth,
		},
	}); err != nil {
		return fmt.Errorf("unable to save plan: %w", err)
	}

	l.Info("queueing job and waiting on execution")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:      runnerJob.ID,
		RunnerID:   install.RunnerID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		return fmt.Errorf("unable to execute job: %w", err)
	}

	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		return fmt.Errorf("unable to sync secrets: %w", err)
	}

	return nil
}

func (w *Workflows) getRoleForSecretSync(
	l *zap.Logger,
	appConfig *app.AppConfig,
	runtimeRole string,
	stack *app.InstallStack,
	installState *state.State,
) (*operationroles.RoleSelection, error) {
	var entityRoles map[app.OperationType]string
	if appConfig.SecretsConfig.Role != "" {
		entityRoles = map[app.OperationType]string{
			app.OperationSync: appConfig.SecretsConfig.Role,
		}
	}

	selectionCtx := &operationroles.SelectionContext{
		Operation:     app.OperationSync,
		PrincipalType: principal.TypeSecret,
		PrincipalName: "",
		RuntimeRole:   runtimeRole,
		EntityRoles:   entityRoles,
		MatrixRules:   appConfig.OperationRoleConfig.Rules,
		DefaultRole:   appConfig.PermissionsConfig.MaintenanceRole.Name,
		AppConfig:     appConfig,
		StackOutputs:  &stack.InstallStackOutputs,
		InstallState:  installState,
	}

	roleSelection, err := operationroles.SelectRole(selectionCtx, l)
	if err != nil {
		l.Warn("dynamic role selection failed, falling back to default role",
			zap.Error(err),
			zap.String("default_role", selectionCtx.DefaultRole),
		)

		var fallbackErr error
		roleSelection, fallbackErr = operationroles.GetDefaultRoleSelection(selectionCtx)
		if fallbackErr != nil {
			return nil, fmt.Errorf("unable to get default role: %w", fallbackErr)
		}

		l.Warn("using default role for secret sync",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	return roleSelection, nil
}

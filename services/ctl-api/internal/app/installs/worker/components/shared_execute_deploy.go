package components

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	pkgplan "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

func (w *Workflows) execPlan(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, stepID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating deploy plan")
	l.Info("planning and executing deploy")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return fmt.Errorf("unable to get build: %w", err)
	}

	appConfig, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get app config")
		return fmt.Errorf("unable to get app config: %w", err)
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, build.ComponentID)
	if err != nil {
		return errors.Wrap(err, "unable to get component")
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	op := app.RunnerJobOperationTypeCreateTeardownPlan
	if installDeploy.Type == app.InstallDeployTypeApply {
		op = app.RunnerJobOperationTypeCreateApplyPlan
	}

	jobTyp := build.ComponentConfigConnection.Type.DeployPlanJobType()

	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:        install.RunnerGroup.Runners[0].ID,
		DeployID:        installDeploy.ID,
		Op:              op,
		Type:            jobTyp,
		LogStreamID:     logStreamID,
		TimeoutDuration: build.ComponentConfigConnection.GetDeployTimeout(),
		Metadata: map[string]string{
			"install_id":           install.ID,
			"deploy_id":            installDeploy.ID,
			"install_component_id": installDeploy.InstallComponentID,
			"component_id":         build.ComponentConfigConnection.ComponentID,
			"component_name":       build.ComponentConfigConnection.Component.Name,
			"operation":            string(op),
		},
	})
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, install.ID)
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get install stack")
		return errors.Wrap(err, "unable to get install stack")
	}

	plan, err := pkgplan.AwaitCreateDeployPlan(ctx, &pkgplan.CreateDeployPlanRequest{
		InstallDeployID: installDeploy.ID,
		InstallID:       install.ID,
		WorkflowID:      fmt.Sprintf("%s-create-deploy-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create deploy plan")
		return errors.Wrap(err, "unable to create deploy plan")
	}

	planJSON, err := json.Marshal(plan)
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create json from deploy plan")
		return errors.Wrap(err, "unable to create json from plan")
	}

	// Get install state for role name rendering
	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get install state")
		return fmt.Errorf("unable to get install state: %w", err)
	}

	compositePlan := plantypes.CompositePlan{
		DeployPlan: plan,
	}

	roleSelection, operation, err := w.getRoleForDeploy(l, appConfig, installDeploy, build, comp, stack, installState)
	if err != nil {
		l.Error("unable to evaluate role for component operation", zap.Error(err))
		w.updateDeployStatusWithoutStatusSync(
			ctx,
			installDeploy.ID,
			app.InstallDeployStatusError,
			"unable to select role",
		)
		return errors.Wrap(err, "unable to evaluate role for component deploy")
	}

	l.Info("selected role for component deploy",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("component", comp.Name),
		zap.String("operation", string(operation)),
	)

	planAuth, err := pkgplan.CreatePlanAuth(
		stack.InstallStackOutputs,
		roleSelection.RoleARN,
		fmt.Sprintf("install-deploy-%s", installDeploy.ID),
	)
	if err != nil {
		l.Error("unable to build plan auth for component operation", zap.Error(err))
		w.updateDeployStatusWithoutStatusSync(
			ctx,
			installDeploy.ID,
			app.InstallDeployStatusError,
			"unable to create auth config",
		)
		return fmt.Errorf("unable to create plan auth: %w", err)
	}

	compositePlan.Auth = planAuth

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: compositePlan,
	}); err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	planJSON = nil
	plan = nil

	w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "creating plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to execute runner job")
		l.Error("job did not succeed", zap.Error(err))
		return fmt.Errorf("unable to get install: %w", err)
	}

	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job")
	}

	var approvalTyp app.WorkflowStepApprovalType
	switch comp.Type {
	case app.ComponentTypeHelmChart:
		approvalTyp = app.HelmApprovalApprovalType
	case app.ComponentTypeTerraformModule:
		approvalTyp = app.TerraformPlanApprovalType
	case app.ComponentTypeKubernetesManifest:
		approvalTyp = app.KubernetesManifestApprovalType
	default:
		approvalTyp = app.NoopApprovalType
	}

	// all component types require a plan EXCEPT fro docker builds
	if job.Execution.Result == nil || (len(job.Execution.Result.Contents) < 1 && len(job.Execution.Result.ContentsGzip) < 1) {
		if runnerJob.Type != app.RunnerJobTypeJobNOOPDeploy {
			return errors.New("no plan returned from job")
		}
	}

	var contentsDisplay string
	if runnerJob.Type == app.RunnerJobTypeJobNOOPDeploy {
		l.Debug("job is a noop - no plan is expected")
		contentsDisplay = "{\"op\": \"noop\"}"
	} else {
		contentsDisplay, err = job.Execution.Result.GetContentsDisplayString()
		if err != nil {
			return errors.Wrap(err, "unable to get content display")
		}
	}

	if _, err := activities.AwaitCreateStepApproval(ctx, &activities.CreateStepApprovalRequest{
		OwnerID:     installDeploy.ID,
		OwnerType:   "install_deploys",
		RunnerJobID: job.ID,
		StepID:      stepID,
		Plan:        contentsDisplay,
		Type:        approvalTyp,
	}); err != nil {
		return errors.Wrap(err, "unable to create approval")
	}

	return nil
}

func (w *Workflows) getRoleForDeploy(
	l *zap.Logger,
	appConfig *app.AppConfig,
	installDeploy *app.InstallDeploy,
	build *app.ComponentBuild,
	comp *app.Component,
	stack *app.InstallStack,
	installState *state.State,
) (*operationroles.RoleSelection, app.OperationType, error) {
	operation := app.OperationDeploy
	if installDeploy.Type == app.InstallDeployTypeTeardown {
		operation = app.OperationTeardown
	}

	selectionCtx := &operationroles.SelectionContext{
		Operation:     operation,
		PrincipalType: principal.TypeComponent,
		PrincipalName: comp.Name,
		RuntimeRole:   installDeploy.Role,
		EntityRoles: operationroles.EntityOperationRoleMapFromHstore(
			build.ComponentConfigConnection.OperationRoles,
		),
		MatrixRules:  appConfig.OperationRoleConfig.Rules,
		DefaultRole:  appConfig.PermissionsConfig.MaintenanceRole.Name,
		AppConfig:    appConfig,
		StackOutputs: &stack.InstallStackOutputs,
		InstallState: installState,
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
			return nil, "", fmt.Errorf("unable to get default role: %w", fallbackErr)
		}

		l.Warn("using default role for component deploy",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	return roleSelection, operation, nil
}

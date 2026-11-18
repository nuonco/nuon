package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) workflowResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Workflows.In(s.db, orgID)
}

func (s *service) workflowTarget(ctx *gin.Context) (permissions.Target, error) {
	resources, err := s.workflowResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	id := ctx.Param("workflow_id")
	if id == "" {
		id = ctx.Param("install_workflow_id")
	}
	return resources.Resolve(ctx.Request.Context(), id)
}

func (s *service) workflowContinuation(operation permissions.PermissionID) func(*gin.Context) ([]authz.Requirement, error) {
	return func(ctx *gin.Context) ([]authz.Requirement, error) {
		target, err := s.workflowTarget(ctx)
		if err != nil {
			return nil, err
		}
		return s.workflowRequirements(ctx, target, operation)
	}
}

func (s *service) workflowApprovalRequirements(ctx *gin.Context) ([]authz.Requirement, error) {
	var req CreateWorkflowStepApprovalResponseRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}
	requirements, err := s.workflowContinuation(permissions.WorkflowApprove)(ctx)
	if err != nil {
		return nil, err
	}
	if req.ResponseType == app.WorkflowStepApprovalResponseTypeRetryPlan {
		requirements = append(requirements, authz.Requirement{Operation: permissions.WorkflowRetry, Target: requirements[0].Target})
	}
	return requirements, nil
}

func (s *service) installWorkflowRetryRequirements(ctx *gin.Context) ([]authz.Requirement, error) {
	var req RetryWorkflowRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}
	resources, err := s.workflowResources(ctx)
	if err != nil {
		return nil, err
	}
	target, err := resources.Resolve(ctx.Request.Context(), req.WorkflowID)
	if err != nil {
		return nil, err
	}
	install, err := s.installTarget(ctx)
	if err != nil {
		return nil, err
	}
	if target.InstallID != install.ID {
		return nil, gorm.ErrRecordNotFound
	}
	return s.workflowRequirements(ctx, target, permissions.WorkflowRetry)
}

func (s *service) workflowRequirements(ctx context.Context, target permissions.Target, operation permissions.PermissionID) ([]authz.Requirement, error) {
	var workflow app.Workflow
	if err := s.db.WithContext(ctx).Where(&app.Workflow{ID: target.ID, OrgID: target.OrgID}).Take(&workflow).Error; err != nil {
		return nil, err
	}
	if target.InstallID == "" && target.AppID != "" {
		switch workflow.OwnerType {
		case "apps", "app_branches", "app_branch_configs":
			switch workflow.Type {
			case app.WorkflowTypeAppConfigBuild, app.WorkflowTypeAppInstallSync, app.WorkflowTypeAppBranchesRun,
				app.WorkflowTypeAppBranchesConfigRepoUpdate, app.WorkflowTypeAppBranchesComponentRepoUpdate:
				return []authz.Requirement{{Operation: operation, Target: target}, {Operation: permissions.AppAutomationManage, Target: target}}, nil
			}
		}
		return nil, gorm.ErrRecordNotFound
	}
	if workflow.OwnerType != "installs" || workflow.OwnerID != target.InstallID || target.InstallID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var underlying permissions.PermissionID
	switch workflow.Type {
	case app.WorkflowTypeDeprovision:
		underlying = permissions.InstallDelete
	case app.WorkflowTypeProvision, app.WorkflowTypeReprovision, app.WorkflowTypeReprovisionStack, app.WorkflowTypeReprovisionSandbox, app.WorkflowTypeDriftRunReprovisionSandbox:
		underlying = permissions.InstallProvision
	case app.WorkflowTypeManualDeploy, app.WorkflowTypeDeployComponents, app.WorkflowTypeRecoverHelmRelease, app.WorkflowTypeDriftRun:
		underlying = permissions.InstallDeploy
	case app.WorkflowTypeDeprovisionSandbox, app.WorkflowTypeTeardownComponent, app.WorkflowTypeTeardownComponents:
		underlying = permissions.InstallTeardown
	case app.WorkflowTypeInputUpdate:
		underlying = permissions.InstallInputsUpdate
	case app.WorkflowTypeAppBranchConfigUpdate, app.WorkflowTypeAppInstallSync, app.WorkflowTypeComponentEnabled, app.WorkflowTypeComponentDisabled:
		underlying = permissions.InstallConfigUpdate
	case app.WorkflowTypeSyncSecrets:
		underlying = permissions.InstallSecretsSync
	case app.WorkflowTypeActionWorkflowRun:
		actionOperation, err := s.actionWorkflowPermission(ctx, &workflow, target)
		if err != nil {
			return nil, err
		}
		underlying = actionOperation
	case app.WorkflowTypeRunbookRun:
		if err := s.validateRunbookContinuation(ctx, &workflow, target); err != nil {
			return nil, err
		}
		underlying = permissions.RunbookExecute
	default:
		return nil, gorm.ErrRecordNotFound
	}
	if resolveDrift := workflow.Metadata[app.WorkflowMetadataKeyResolveDrift]; resolveDrift != nil && *resolveDrift == "true" {
		underlying = permissions.InstallDriftResolve
	}
	install := permissions.Target{ResolvedTarget: permissions.ResolvedTarget{
		Scope: permissions.Scope{Kind: permissions.KindInstall, OrgID: target.OrgID, ID: target.InstallID},
		AppID: target.AppID, InstallID: target.InstallID,
	}}
	return []authz.Requirement{{Operation: operation, Target: target}, {Operation: underlying, Target: install}}, nil
}

func (s *service) actionWorkflowPermission(ctx context.Context, workflow *app.Workflow, target permissions.Target) (permissions.PermissionID, error) {
	metadata := func(key string) string {
		if value := workflow.Metadata[key]; value != nil {
			return *value
		}
		return ""
	}
	var runs []app.InstallActionWorkflowRun
	if err := s.db.WithContext(ctx).Where(app.InstallActionWorkflowRun{InstallWorkflowID: &workflow.ID}).Find(&runs).Error; err != nil {
		return "", err
	}
	adhoc := metadata("adhoc_action") != "" || metadata("adhoc_action_run_id") != ""
	if adhoc {
		if metadata("adhoc_action") != "true" || metadata("trigger_type") != "adhoc" || metadata("install_action_workflow_id") != "" || len(runs) != 1 || runs[0].ID != metadata("adhoc_action_run_id") {
			return "", gorm.ErrRecordNotFound
		}
	} else {
		if metadata("install_action_workflow_id") == "" {
			return "", gorm.ErrRecordNotFound
		}
		resources, err := actionaccess.InstallActions.In(s.db, target.OrgID)
		if err != nil {
			return "", err
		}
		action, err := resources.Resolve(ctx, metadata("install_action_workflow_id"))
		if err != nil {
			return "", err
		}
		if action.InstallID != target.InstallID || action.AppID != target.AppID {
			return "", gorm.ErrRecordNotFound
		}
	}
	for _, run := range runs {
		if run.OrgID != target.OrgID || run.InstallID != target.InstallID {
			return "", gorm.ErrRecordNotFound
		}
		var notebookRuns int64
		if err := s.db.WithContext(ctx).Model(&app.NotebookCellRun{}).
			Where(app.NotebookCellRun{InstallActionWorkflowRunID: run.ID}).Count(&notebookRuns).Error; err != nil {
			return "", err
		}
		if notebookRuns != 0 {
			return "", gorm.ErrRecordNotFound
		}
		if adhoc {
			if run.TriggerType != app.ActionWorkflowTriggerTypeAdHoc || run.TriggeredByType != "account" || run.InstallActionWorkflowID.Valid || run.ActionWorkflowConfigID.Valid {
				return "", gorm.ErrRecordNotFound
			}
		} else {
			if !run.InstallActionWorkflowID.Valid || run.InstallActionWorkflowID.String != metadata("install_action_workflow_id") || !run.ActionWorkflowConfigID.Valid || run.TriggerType == app.ActionWorkflowTriggerTypeAdHoc {
				return "", gorm.ErrRecordNotFound
			}
			var installed app.InstallActionWorkflow
			if err := s.db.WithContext(ctx).Where(app.InstallActionWorkflow{ID: run.InstallActionWorkflowID.String, OrgID: target.OrgID, InstallID: target.InstallID}).Take(&installed).Error; err != nil {
				return "", err
			}
			var config app.ActionWorkflowConfig
			if err := s.db.WithContext(ctx).Where(app.ActionWorkflowConfig{ID: run.ActionWorkflowConfigID.String, OrgID: target.OrgID, AppID: target.AppID, ActionWorkflowID: installed.ActionWorkflowID}).Take(&config).Error; err != nil {
				return "", err
			}
		}
	}
	if adhoc {
		return permissions.ActionAdhocExecute, nil
	}
	return permissions.ActionExecute, nil
}

func (s *service) validateRunbookContinuation(ctx context.Context, workflow *app.Workflow, target permissions.Target) error {
	metadata := func(key string) string {
		if value := workflow.Metadata[key]; value != nil {
			return *value
		}
		return ""
	}
	if metadata("install_runbook_run_id") == "" || metadata("install_runbook_id") == "" || metadata("runbook_config_id") == "" || metadata("install_id") != target.InstallID {
		return gorm.ErrRecordNotFound
	}
	var run app.InstallRunbookRun
	if err := s.db.WithContext(ctx).Where(app.InstallRunbookRun{
		ID: metadata("install_runbook_run_id"), OrgID: target.OrgID, InstallID: target.InstallID,
		InstallWorkflowID: &workflow.ID, InstallRunbookID: metadata("install_runbook_id"), RunbookConfigID: metadata("runbook_config_id"),
	}).Take(&run).Error; err != nil {
		return err
	}
	var installed app.InstallRunbook
	if err := s.db.WithContext(ctx).Where(app.InstallRunbook{ID: run.InstallRunbookID, OrgID: target.OrgID, InstallID: target.InstallID}).Take(&installed).Error; err != nil {
		return err
	}
	var config app.RunbookConfig
	if err := s.db.WithContext(ctx).Where(app.RunbookConfig{ID: run.RunbookConfigID, OrgID: target.OrgID, AppID: target.AppID, RunbookID: installed.RunbookID}).Preload("Steps").Take(&config).Error; err != nil {
		return err
	}
	for _, step := range config.Steps {
		if step.Type == app.RunbookStepTypeWaitForEvent {
			authority, err := authz.AuthorityFromContext(ctx)
			if err != nil {
				return err
			}
			if step.TriggerID == "" || !authority.Check(permissions.AppAutomationManage, target).Matched {
				return gorm.ErrRecordNotFound
			}
			var trigger app.Trigger
			if err := s.db.WithContext(ctx).Where(app.Trigger{ID: step.TriggerID, OrgID: target.OrgID}).First(&trigger).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

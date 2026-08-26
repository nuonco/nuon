package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// @ID						CreateInstallActionWorkflowRun
// @Summary					create an action workflow run for an install
// @Description.markdown	create_install_action_workflow_run.md
// @Tags					actions
// @Accept					json
// @Param					install_id	path	string									true	"install ID"
// @Param					req			body	CreateInstallActionWorkflowRunRequest	true	"Input"
// @Produce					json
// @Security				APIKey && OrgID
// @Deprecated 				true
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					409	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	app.WorkflowResponse
// @Router					/v1/installs/{install_id}/action-workflows/runs [post]
func (s *service) CreateInstallActionWorkflowRun(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallActionWorkflowRunRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	result, err := s.createInstallActionWorkflowRun(ctx, installID, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: result.WorkflowID})
}

type createInstallActionWorkflowRunResult struct {
	WorkflowID             string
	ActionWorkflowID       string
	ActionWorkflowConfigID string
}

func (s *service) createInstallActionWorkflowRun(ctx context.Context, installID string, req CreateInstallActionWorkflowRunRequest) (*createInstallActionWorkflowRunResult, error) {
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	awc, err := s.actionsHelpers.GetActionWorkflowConfigByID(ctx, req.ActionWorkFlowConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to get action workflow config: %w", err)
	}

	// Callers pass whichever config they last read, usually the app's newest rather
	// than the install's pinned one, so re-resolve instead of rejecting.
	if awc.AppConfigID != install.AppConfigID {
		pinned, err := s.actionsHelpers.GetActionWorkflowConfig(ctx, awc.ActionWorkflowID, install.AppConfigID)
		if err != nil {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("action is not in the install's app config version: %w", err),
				Description: "this action is not in the install's app config version",
			}
		}
		awc = pinned
	}

	if !awc.WorkflowConfigCanTriggerManually() {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("manual trigger is not allowed"),
			Description: "please update action config to allow manual triggering",
		}
	}

	installActionWorkflow, err := s.getInstallActionWorkflow(ctx, installID, awc.ActionWorkflowID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install action workflow")
	}

	prependRunEnvVars := PrependRunEnvPrefix(req.RunEnvVars)

	workflowMetadata := make(map[string]string)
	workflowMetadata["install_action_workflow_id"] = installActionWorkflow.ID
	workflowMetadata["install_action_workflow_name"] = installActionWorkflow.ActionWorkflow.Name

	accountID := keys.CreatedByIDFromContext(ctx)
	if accountID == "" {
		account, err := cctx.AccountFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get account from context: %w", err)
		}
		accountID = account.ID
	}
	workflowMetadata["triggerred_by_id"] = accountID

	for k, v := range prependRunEnvVars {
		workflowMetadata[k] = v
	}

	workflow, err := s.installHelpers.CreateWorkflowWithRole(ctx,
		installActionWorkflow.InstallID,
		app.WorkflowTypeActionWorkflowRun,
		workflowMetadata,
		false,
		req.Role,
	)
	if err != nil {
		return nil, err
	}

	queueID, err := s.getInstallActionWorkflowsQueueID(ctx, installActionWorkflow.InstallID)
	if err != nil {
		return nil, err
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
		WorkflowID: workflow.ID,
	}, workflow.ID, "install_workflows"); err != nil {
		return nil, fmt.Errorf("enqueue signal: %w", err)
	}

	return &createInstallActionWorkflowRunResult{
		WorkflowID:             workflow.ID,
		ActionWorkflowID:       awc.ActionWorkflowID,
		ActionWorkflowConfigID: awc.ID,
	}, nil
}

// PrependRunEnvPrefix modifies the keys in the provided RunEnvVars map
// by prepending "RUNENV_" to each key.
func PrependRunEnvPrefix(runEnvVars map[string]string) map[string]string {
	result := make(map[string]string, len(runEnvVars))

	for key, value := range runEnvVars {
		newKey := "RUNENV_" + key
		result[newKey] = value
	}

	return result
}

func (s *service) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	var install app.Install
	res := s.db.WithContext(ctx).First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install")
	}

	return &install, nil
}

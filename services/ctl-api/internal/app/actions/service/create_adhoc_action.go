package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAdHocActionRequest struct {
	InlineContents   string            `json:"inline_contents" validate:"required_without=Command"`
	Command          string            `json:"command" validate:"required_without=InlineContents"`
	EnvVars          map[string]string `json:"env_vars"`
	Timeout          int               `json:"timeout,omitempty" validate:"omitempty,min=1,max=3600"`
	Name             string            `json:"name" validate:"max=255"`
	Role             string            `json:"role"`
	EnableKubeConfig *bool             `json:"enable_kube_config" extensions:"x-nullable"`
}

func (c *CreateAdHocActionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if c.InlineContents != "" && c.Command != "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("provide either inline_contents or command, not both"),
			Description: "invalid request input",
		}
	}
	if c.InlineContents == "" && c.Command == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("either inline_contents or command is required"),
			Description: "invalid request input",
		}
	}

	if c.Timeout == 0 {
		c.Timeout = 300
	}

	return nil
}

type CreateAdHocActionResponse struct {
	ID                string                             `json:"id"`
	InstallID         string                             `json:"install_id"`
	Status            app.InstallActionWorkflowRunStatus `json:"status"`
	StatusDescription string                             `json:"status_description"`
	TriggerType       app.ActionWorkflowTriggerType      `json:"trigger_type"`
	CreatedAt         time.Time                          `json:"created_at"`
	WorkflowID        string                             `json:"workflow_id"`
}

// @ID                       CreateAdHocAction
// @Summary                  create an adhoc action run for an install
// @Description.markdown     create_adhoc_action.md
// @Tags                     actions
// @Accept                   json
// @Param                    install_id  path    string                      true    "install ID"
// @Param                    req         body    CreateAdHocActionRequest    true    "Input"
// @Produce                  json
// @Security                 APIKey
// @Security                 OrgID
// @Failure                  400 {object} stderr.ErrResponse
// @Failure                  401 {object} stderr.ErrResponse
// @Failure                  403 {object} stderr.ErrResponse
// @Failure                  404 {object} stderr.ErrResponse
// @Failure                  500 {object} stderr.ErrResponse
// @Success                  201 {object} CreateAdHocActionResponse
// @Router                   /v1/installs/{install_id}/actions/adhoc-run [post]
func (s *service) CreateAdHocAction(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateAdHocActionRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         err,
			Description: "invalid request body",
		})
		return
	}

	if err := req.Validate(s.v); err != nil {
		ctx.Error(err)
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         err,
			Description: "install not found",
		})
		return
	}

	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	run, err := s.createAdHocActionRun(ctx, install, account.ID, &req)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         err,
			Description: "failed to create adhoc action run",
		})
		return
	}

	actionName := req.Name
	if actionName == "" {
		if req.InlineContents != "" {
			actionName = "Adhoc script"
		} else {
			actionName = "Adhoc command"
		}
	}

	prependRunEnvVars := PrependRunEnvPrefix(req.EnvVars)
	prependRunEnvVars["adhoc_action_run_id"] = run.ID
	prependRunEnvVars["triggered_by_id"] = account.ID
	prependRunEnvVars["trigger_type"] = "adhoc"
	prependRunEnvVars["install_action_workflow_name"] = actionName
	prependRunEnvVars["adhoc_action"] = "true"

	workflow, err := s.installHelpers.CreateWorkflowWithRole(ctx,
		install.ID,
		app.WorkflowTypeActionWorkflowRun,
		prependRunEnvVars,
		false,
		req.Role,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	run.InstallWorkflowID = &workflow.ID
	if err := s.db.WithContext(ctx).Save(run).Error; err != nil {
		ctx.Error(err)
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
		queueID, err := s.getInstallActionWorkflowsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
			WorkflowID: workflow.ID,
		}, workflow.ID, "install_workflows"); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type:              signals.OperationExecuteFlow,
			InstallWorkflowID: workflow.ID,
		})
	}

	ctx.JSON(http.StatusCreated, CreateAdHocActionResponse{
		ID:                run.ID,
		InstallID:         install.ID,
		Status:            run.Status,
		StatusDescription: run.StatusDescription,
		TriggerType:       app.ActionWorkflowTriggerTypeAdHoc,
		CreatedAt:         run.CreatedAt,
		WorkflowID:        workflow.ID,
	})
}

func (s *service) createAdHocActionRun(
	ctx context.Context,
	install *app.Install,
	accountID string,
	req *CreateAdHocActionRequest,
) (*app.InstallActionWorkflowRun, error) {
	stepConfig := app.ActionWorkflowStepConfig{
		InlineContents: req.InlineContents,
		Command:        req.Command,
		EnvVars:        dbgenerics.ToHstore(req.EnvVars),
		Name:           req.Name,
		Idx:            0,
	}

	if stepConfig.Name == "" {
		if req.InlineContents != "" {
			stepConfig.Name = "Adhoc script"
		} else {
			stepConfig.Name = "Adhoc command"
		}
	}

	adHocConfig := app.AdHocStepConfig(stepConfig)
	runStep := app.InstallActionWorkflowRunStep{
		Status:      app.InstallActionWorkflowRunStepStatusPending,
		AdHocConfig: &adHocConfig,
	}

	defaultEnableKubeConfig := true
	enableKubeConfig := generics.NewNullBoolFromPtr(&defaultEnableKubeConfig)
	if req.EnableKubeConfig != nil {
		enableKubeConfig = generics.NewNullBoolFromPtr(req.EnableKubeConfig)
	}

	run := app.InstallActionWorkflowRun{
		InstallID:         install.ID,
		TriggerType:       app.ActionWorkflowTriggerTypeAdHoc,
		TriggeredByID:     accountID,
		TriggeredByType:   "account",
		Status:            app.InstallActionRunStatusQueued,
		StatusDescription: "Queued for execution",
		Steps:             []app.InstallActionWorkflowRunStep{runStep},
		RunEnvVars:        dbgenerics.ToHstore(req.EnvVars),
		Timeout:           time.Duration(req.Timeout) * time.Second,
		Role:              req.Role,
		EnableKubeConfig:  enableKubeConfig,
	}

	if err := s.db.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, err
	}

	return &run, nil
}

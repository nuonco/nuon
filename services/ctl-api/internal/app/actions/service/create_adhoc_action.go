package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/apiidem"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAdHocActionRequest struct {
	// RequestID is an optional idempotency key. The same id and body returns the original run. A different body returns 409.
	RequestID        string            `json:"request_id,omitempty" validate:"omitempty,max=255"`
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
// @Failure                  409 {object} stderr.ErrResponse
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

	if req.RequestID != "" {
		resp, err := s.createIdempotentAdHocAction(ctx, install, account.ID, &req)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusCreated, resp)
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

	queueID, err := s.getInstallActionWorkflowsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
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
	run, err := s.insertAdHocActionRun(ctx, s.db, install, accountID, req)
	if err != nil {
		return nil, err
	}
	s.emitAdHocActionAudit(ctx, install.ID, run.ID, accountID)
	return run, nil
}

func (s *service) createIdempotentAdHocAction(ctx context.Context, install *app.Install, accountID string, req *CreateAdHocActionRequest) (*CreateAdHocActionResponse, error) {
	hashReq := *req
	hashReq.RequestID = ""
	hash, err := apiidem.Hash(hashReq)
	if err != nil {
		return nil, err
	}

	actionName := req.Name
	if actionName == "" {
		if req.InlineContents != "" {
			actionName = "Adhoc script"
		} else {
			actionName = "Adhoc command"
		}
	}

	var createdRun *app.InstallActionWorkflowRun
	workflow, created, err := s.installHelpers.RunIdempotentInstallWorkflow(ctx, installhelpers.IdempotentInstallWorkflowRequest{
		InstallID:             install.ID,
		WorkflowType:          app.WorkflowTypeActionWorkflowRun,
		Role:                  req.Role,
		RequestID:             req.RequestID,
		RequestHash:           hash,
		Operation:             "actions-adhoc-run",
		QueueName:             installhelpers.InstallActionWorkflowsQueueName,
		SkipAppConfigConflict: true,
	}, &installhelpers.IdempotentInstallWorkflowHooks{
		Prepare: func(tx *gorm.DB) (map[string]string, error) {
			run, err := s.insertAdHocActionRun(ctx, tx, install, accountID, req)
			if err != nil {
				return nil, err
			}
			createdRun = run
			metadata := PrependRunEnvPrefix(req.EnvVars)
			metadata["adhoc_action_run_id"] = run.ID
			metadata["triggered_by_id"] = accountID
			metadata["trigger_type"] = "adhoc"
			metadata["install_action_workflow_name"] = actionName
			metadata["adhoc_action"] = "true"
			return metadata, nil
		},
		AfterCreate: func(tx *gorm.DB, wf *app.Workflow) error {
			if createdRun == nil {
				return fmt.Errorf("adhoc action run was not created")
			}
			createdRun.InstallWorkflowID = &wf.ID
			return tx.Model(createdRun).Update("install_workflow_id", wf.ID).Error
		},
	})
	if err != nil {
		return nil, err
	}
	if created && createdRun != nil {
		s.emitAdHocActionAudit(ctx, install.ID, createdRun.ID, accountID)
	}

	run := createdRun
	if run == nil {
		runID := generics.FromPtrStr(workflow.Metadata["adhoc_action_run_id"])
		run = &app.InstallActionWorkflowRun{}
		if err := s.db.WithContext(ctx).Where(app.InstallActionWorkflowRun{ID: runID}).First(run).Error; err != nil {
			return nil, err
		}
	}

	return &CreateAdHocActionResponse{
		ID:                run.ID,
		InstallID:         install.ID,
		Status:            run.Status,
		StatusDescription: run.StatusDescription,
		TriggerType:       app.ActionWorkflowTriggerTypeAdHoc,
		CreatedAt:         run.CreatedAt,
		WorkflowID:        workflow.ID,
	}, nil
}

func (s *service) insertAdHocActionRun(
	ctx context.Context,
	db *gorm.DB,
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

	if err := db.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *service) emitAdHocActionAudit(ctx context.Context, installID, runID, accountID string) {
	s.audit.Emit(ctx, audit.Event{
		Type:        audit.EventInstallActionWorkflowRun,
		Message:     "adhoc action run created",
		Outcome:     audit.OutcomeStarted,
		InstallID:   installID,
		SubjectID:   runID,
		SubjectType: "install_action_workflow_runs",
		Attrs: map[string]string{
			"action_workflow_run.id": runID,
			"trigger.type":           string(app.ActionWorkflowTriggerTypeAdHoc),
			"triggered_by.id":        accountID,
			"triggered_by.type":      "account",
		},
	})
}

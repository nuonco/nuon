package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateRunbookConfigRequest struct {
	AppConfigID *string                           `json:"app_config_id"`
	Readme      string                            `json:"readme"`
	Steps       []*CreateRunbookStepConfigRequest `json:"steps" validate:"required"`
}

type CreateRunbookStepConfigRequest struct {
	Name                 string `json:"name" validate:"required"`
	Type                 string `json:"type" validate:"required"`
	Idx                  int64  `json:"idx"`
	ComponentName        string `json:"component_name,omitempty"`
	DeployDependents     bool   `json:"deploy_dependents,omitempty"`
	TearDownDependents   bool   `json:"tear_down_dependents,omitempty"`
	SkipComponentDeploys bool   `json:"skip_component_deploys,omitempty"`
	// Legacy alias for DeployDependents — accepted to keep older API clients working.
	DeployDependenciesLegacy bool              `json:"deploy_dependencies,omitempty" swaggerignore:"true"`
	ActionName               string            `json:"action_name,omitempty"`
	Command                  string            `json:"command,omitempty"`
	InlineContents           string            `json:"inline_contents,omitempty"`
	EnvVars                  map[string]string `json:"env_vars,omitempty"`
	Timeout                  int64             `json:"timeout,omitempty"`
	Role                     string            `json:"role,omitempty"`
}

// @ID				CreateRunbookConfig
// @Summary		create a runbook config
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id		path	string						true	"app ID"
// @Param			runbook_id	path	string						true	"runbook ID"
// @Param			req			body	CreateRunbookConfigRequest	true	"Input"
// @Success		201			{object}	app.RunbookConfig
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/apps/{app_id}/runbooks/{runbook_id}/configs [post]
func (s *service) CreateRunbookConfig(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	runbookID := ctx.Param("runbook_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Look up the runbook to get the real app ID (SDK may pass "_" as app_id).
	var runbook app.Runbook
	if res := s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: org.ID}).
		Where("id = ?", runbookID).
		First(&runbook); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runbook: %w", res.Error))
		return
	}

	var req CreateRunbookConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	appConfigID := ""
	if req.AppConfigID != nil {
		appConfigID = *req.AppConfigID
	}

	steps := make([]app.RunbookStepConfig, 0, len(req.Steps))
	for idx, stepReq := range req.Steps {
		stepType := app.RunbookStepType(stepReq.Type)
		// Canonicalize the legacy "deploy" step type to "component_deploy".
		if stepType == app.RunbookStepTypeDeployLegacy {
			stepType = app.RunbookStepTypeComponentDeploy
		}
		switch stepType {
		case app.RunbookStepTypeComponentDeploy,
			app.RunbookStepTypeComponentTearDown,
			app.RunbookStepTypeAction,
			app.RunbookStepTypeSandboxReprovision,
			app.RunbookStepTypeSandboxDeprovision:
		default:
			ctx.Error(fmt.Errorf("invalid step type %q for step %s", stepReq.Type, stepReq.Name))
			return
		}

		envVars := pgtype.Hstore{}
		for k, v := range stepReq.EnvVars {
			envVars[k] = &v
		}

		stepCfg := app.RunbookStepConfig{
			Idx:                  idx,
			Name:                 stepReq.Name,
			Type:                 stepType,
			ComponentName:        stepReq.ComponentName,
			DeployDependents:     stepReq.DeployDependents || stepReq.DeployDependenciesLegacy,
			TearDownDependents:   stepReq.TearDownDependents,
			SkipComponentDeploys: stepReq.SkipComponentDeploys,
			Command:              stepReq.Command,
			InlineContents:       stepReq.InlineContents,
			EnvVars:              envVars,
			Timeout:              time.Duration(stepReq.Timeout),
			Role:                 stepReq.Role,
		}

		// Resolve action_name to ActionWorkflowID
		if stepReq.ActionName != "" {
			var aw app.ActionWorkflow
			if err := s.db.WithContext(ctx).
				Where(app.ActionWorkflow{AppID: runbook.AppID, Name: stepReq.ActionName}).
				First(&aw).Error; err != nil {
				ctx.Error(fmt.Errorf("unable to find action %q for step %s: %w", stepReq.ActionName, stepReq.Name, err))
				return
			}
			stepCfg.ActionWorkflowID = generics.NewNullString(aw.ID)
		}

		steps = append(steps, stepCfg)
	}

	rbcfg := app.RunbookConfig{
		OrgID:       org.ID,
		AppID:       runbook.AppID,
		AppConfigID: appConfigID,
		RunbookID:   runbook.ID,
		Readme:      req.Readme,
		Steps:       steps,
	}

	res := s.db.WithContext(ctx).Create(&rbcfg)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create runbook config: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, rbcfg)
}

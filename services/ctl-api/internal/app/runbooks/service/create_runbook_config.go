package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateRunbookConfigRequest struct {
	AppConfigID *string                           `json:"app_config_id"`
	Readme      string                            `json:"readme"`
	Steps       []*CreateRunbookStepConfigRequest `json:"steps" validate:"required"`
}

type CreateRunbookStepConfigRequest struct {
	Name               string            `json:"name" validate:"required"`
	Type               string            `json:"type" validate:"required"`
	Idx                int64             `json:"idx"`
	ComponentName      string            `json:"component_name,omitempty"`
	DeployDependencies bool              `json:"deploy_dependencies,omitempty"`
	ActionName         string            `json:"action_name,omitempty"`
	Command            string            `json:"command,omitempty"`
	InlineContents     string            `json:"inline_contents,omitempty"`
	EnvVars            map[string]string `json:"env_vars,omitempty"`
	Timeout            int64             `json:"timeout,omitempty"`
	Role               string            `json:"role,omitempty"`
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
// @Router			/v1/apps/{app_id}/runbooks/{runbook_id}/configs [post]
func (s *service) CreateRunbookConfig(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	runbookID := ctx.Param("runbook_id")
	appID := ctx.Param("app_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
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
	for _, stepReq := range req.Steps {
		envVars := pgtype.Hstore{}
		for k, v := range stepReq.EnvVars {
			envVars[k] = &v
		}

		steps = append(steps, app.RunbookStepConfig{
			Idx:                int(stepReq.Idx),
			Name:               stepReq.Name,
			Type:               app.RunbookStepType(stepReq.Type),
			ComponentName:      stepReq.ComponentName,
			DeployDependencies: stepReq.DeployDependencies,
			Command:            stepReq.Command,
			InlineContents:     stepReq.InlineContents,
			EnvVars:            envVars,
			Timeout:            time.Duration(stepReq.Timeout),
			Role:               stepReq.Role,
		})
	}

	rbcfg := app.RunbookConfig{
		OrgID:       org.ID,
		AppID:       appID,
		AppConfigID: appConfigID,
		RunbookID:   runbookID,
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

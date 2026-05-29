package activities

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateAdHocActionRunForRunbookRequest struct {
	InstallID       string        `validate:"required"`
	Command         string        `json:"command"`
	InlineContents  string        `json:"inline_contents"`
	EnvVars         pgtype.Hstore `json:"env_vars"`
	Timeout         time.Duration `json:"timeout"`
	Role            string        `json:"role"`
	TriggeredByID   string        `json:"triggered_by_id"`
	TriggeredByType string        `json:"triggered_by_type"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateAdHocActionRunForRunbook(ctx context.Context, req CreateAdHocActionRunForRunbookRequest) (*app.InstallActionWorkflowRun, error) {
	stepConfig := app.ActionWorkflowStepConfig{
		InlineContents: req.InlineContents,
		Command:        req.Command,
		EnvVars:        req.EnvVars,
		Name:           "runbook-adhoc",
		Idx:            0,
	}

	adHocConfig := app.AdHocStepConfig(stepConfig)
	runStep := app.InstallActionWorkflowRunStep{
		Status:      app.InstallActionWorkflowRunStepStatusPending,
		AdHocConfig: &adHocConfig,
	}

	defaultEnableKubeConfig := true

	run := app.InstallActionWorkflowRun{
		InstallID:         req.InstallID,
		TriggerType:       app.ActionWorkflowTriggerTypeAdHoc,
		TriggeredByID:     req.TriggeredByID,
		TriggeredByType:   req.TriggeredByType,
		Status:            app.InstallActionRunStatusQueued,
		StatusDescription: "Queued for execution",
		Steps:             []app.InstallActionWorkflowRunStep{runStep},
		RunEnvVars:        req.EnvVars,
		Timeout:           req.Timeout,
		Role:              req.Role,
		EnableKubeConfig:  generics.NewNullBoolFromPtr(&defaultEnableKubeConfig),
	}

	if err := a.db.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, err
	}

	return &run, nil
}

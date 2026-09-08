package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateRunbookRunRequest struct {
	Inputs map[string]*string              `json:"inputs,omitempty"`
	Steps  []CreateRunbookRunStepSelection `json:"steps,omitempty"`
	Role   string                          `json:"role,omitempty"`
}

type CreateRunbookRunStepSelection struct {
	StepID  string `json:"step_id" validate:"required"`
	Enabled bool   `json:"enabled"`
}

// @ID				CreateRunbookRun
// @Summary		run a runbook on an install
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			runbook_id	path	string	true	"runbook ID or name"
// @Param			req			body	CreateRunbookRunRequest	false	"Input"
// @Success		201			{object}	app.InstallRunbookRun
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/installs/{install_id}/runbooks/{runbook_id}/runs [post]
func (s *service) CreateRunbookRun(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateRunbookRunRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	triggered, err := s.createRunbookRun(ctx, org.ID, account.ID, ctx.Param("install_id"), ctx.Param("runbook_id"), req)
	if err != nil {
		ctx.Error(err)
		return
	}
	triggered.Run.InstallWorkflow = triggered.Workflow
	ctx.JSON(http.StatusCreated, triggered.Run)
}

func (s *service) createRunbookRun(ctx context.Context, orgID, accountID, installRef, runbookRef string, req CreateRunbookRunRequest) (*runbookshelpers.TriggerRunbookRunResponse, error) {
	var install app.Install
	err := s.db.WithContext(ctx).
		Select("id", "name", "app_id", "app_config_id").
		Where(app.Install{OrgID: orgID}).
		Where(s.db.Where(app.Install{ID: installRef}).Or(app.Install{Name: installRef})).
		First(&install).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	var runbook app.Runbook
	err = s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: orgID, AppID: install.AppID}).
		Where(s.db.Where(app.Runbook{ID: runbookRef}).Or(app.Runbook{Name: runbookRef})).
		First(&runbook).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get runbook: %w", err)
	}

	var installRunbook app.InstallRunbook
	err = s.db.WithContext(ctx).
		Preload("Runbook").
		Where(app.InstallRunbook{OrgID: orgID, InstallID: install.ID, RunbookID: runbook.ID}).
		First(&installRunbook).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get install runbook: %w", err)
	}

	var runbookConfig app.RunbookConfig
	configQuery := s.db.WithContext(ctx).
		Preload("Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Inputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Where(app.RunbookConfig{RunbookID: installRunbook.RunbookID, OrgID: orgID})

	if install.AppConfigID != "" {
		// Never fall back to the newest config: it could run steps the install
		// was not configured with.
		if err := configQuery.Where(app.RunbookConfig{AppConfigID: install.AppConfigID}).First(&runbookConfig).Error; err != nil {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("runbook is not in the install's app config version: %w", err),
				Description: "this runbook is not in the install's app config version",
			}
		}
	} else {
		if err := configQuery.Order("created_at DESC").First(&runbookConfig).Error; err != nil {
			return nil, fmt.Errorf("runbook has no configurations")
		}
	}

	inputsWithDefaults := runbookshelpers.MergeRunbookInputDefaults(&runbookConfig, req.Inputs)
	if err := s.helpers.ValidateRunbookInputs(&runbookConfig, inputsWithDefaults); err != nil {
		return nil, err
	}

	stepSelections, err := buildStepSelections(&runbookConfig, req.Steps)
	if err != nil {
		return nil, err
	}

	inputs := make(map[string]string, len(inputsWithDefaults))
	for name, value := range inputsWithDefaults {
		if value != nil {
			inputs[name] = *value
		}
	}
	triggered, err := s.helpers.TriggerRunbookRun(ctx, runbookshelpers.TriggerRunbookRunRequest{InstallRunbookID: installRunbook.ID, RunbookConfigID: runbookConfig.ID, TriggeredByID: accountID, Inputs: inputs, StepSelections: stepSelections, Role: req.Role})
	if err != nil {
		return nil, err
	}
	return triggered, nil
}

// buildStepSelections validates the supplied step selections against the config's steps
// and returns the persisted selections. Unknown step IDs are rejected; an empty request
// means all steps run. At least one step must remain enabled.
func buildStepSelections(rbConfig *app.RunbookConfig, supplied []CreateRunbookRunStepSelection) ([]app.RunbookStepSelection, error) {
	if len(supplied) == 0 {
		return nil, nil
	}

	stepByID := make(map[string]app.RunbookStepConfig, len(rbConfig.Steps))
	for _, step := range rbConfig.Steps {
		stepByID[step.ID] = step
	}

	selections := make([]app.RunbookStepSelection, 0, len(supplied))
	enabledCount := 0
	for _, sel := range supplied {
		step, ok := stepByID[sel.StepID]
		if !ok {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("step %s does not exist in runbook config", sel.StepID),
				Description: "step " + sel.StepID + " selected for the run does not exist in the runbook",
			}
		}
		if sel.Enabled {
			enabledCount++
		}
		selections = append(selections, app.RunbookStepSelection{
			StepID:  step.ID,
			Name:    step.Name,
			Enabled: sel.Enabled,
		})
	}

	if enabledCount == 0 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("no steps enabled for runbook run"),
			Description: "at least one step must be enabled to run the runbook",
		}
	}

	return selections, nil
}

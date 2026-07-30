package service

import (
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
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	installID := ctx.Param("install_id")
	runbookIDOrName := ctx.Param("runbook_id")
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

	// Find the install to get its app config version
	var install app.Install
	res := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", installID, org.ID).
		First(&install)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", res.Error))
		return
	}

	// Find the install runbook
	var installRunbook app.InstallRunbook
	res = s.db.WithContext(ctx).
		Preload("Runbook").
		Joins("JOIN runbooks ON runbooks.id = install_runbooks.runbook_id AND runbooks.deleted_at = 0").
		Where(app.InstallRunbook{OrgID: org.ID, InstallID: installID}).
		Where("install_runbooks.runbook_id = ? OR runbooks.name = ?", runbookIDOrName, runbookIDOrName).
		First(&installRunbook)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install runbook: %w", res.Error))
		return
	}

	// Find the runbook config matching the install's app config version.
	// Fall back to the latest config if no version-specific config exists.
	var runbookConfig app.RunbookConfig
	configQuery := s.db.WithContext(ctx).
		Preload("Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Inputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Where(app.RunbookConfig{RunbookID: installRunbook.RunbookID, OrgID: org.ID})

	if install.AppConfigID != "" {
		// Try the install's pinned app config first
		if err := configQuery.Where(app.RunbookConfig{AppConfigID: install.AppConfigID}).First(&runbookConfig).Error; err != nil {
			// Fall back to latest config
			if err := s.db.WithContext(ctx).
				Preload("Steps", func(tx *gorm.DB) *gorm.DB {
					return tx.Order("idx ASC")
				}).
				Where(app.RunbookConfig{RunbookID: installRunbook.RunbookID, OrgID: org.ID}).
				Order("created_at DESC").
				First(&runbookConfig).Error; err != nil {
				ctx.Error(fmt.Errorf("runbook has no configurations"))
				return
			}
		}
	} else {
		if err := configQuery.Order("created_at DESC").First(&runbookConfig).Error; err != nil {
			ctx.Error(fmt.Errorf("runbook has no configurations"))
			return
		}
	}

	inputsWithDefaults := runbookshelpers.MergeRunbookInputDefaults(&runbookConfig, req.Inputs)
	if err := s.helpers.ValidateRunbookInputs(&runbookConfig, inputsWithDefaults); err != nil {
		ctx.Error(err)
		return
	}

	stepSelections, err := buildStepSelections(&runbookConfig, req.Steps)
	if err != nil {
		ctx.Error(err)
		return
	}

	inputs := make(map[string]string, len(inputsWithDefaults))
	for name, value := range inputsWithDefaults {
		if value != nil {
			inputs[name] = *value
		}
	}
	triggered, err := s.helpers.TriggerRunbookRun(ctx, runbookshelpers.TriggerRunbookRunRequest{InstallRunbookID: installRunbook.ID, RunbookConfigID: runbookConfig.ID, TriggeredByID: account.ID, Inputs: inputs, StepSelections: stepSelections})
	if err != nil {
		ctx.Error(err)
		return
	}
	triggered.Run.InstallWorkflow = triggered.Workflow
	ctx.JSON(http.StatusCreated, triggered.Run)
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

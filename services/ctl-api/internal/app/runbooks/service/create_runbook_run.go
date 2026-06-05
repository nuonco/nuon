package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// @ID				CreateRunbookRun
// @Summary		run a runbook on an install
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			runbook_id	path	string	true	"runbook ID or name"
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

	// Create the run record
	run := app.InstallRunbookRun{
		OrgID:            org.ID,
		InstallID:        installID,
		InstallRunbookID: installRunbook.ID,
		RunbookConfigID:  runbookConfig.ID,
		Status:           app.InstallRunbookRunStatusQueued,
		TriggeredByID:    account.ID,
	}

	res = s.db.WithContext(ctx).Create(&run)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create runbook run: %w", res.Error))
		return
	}

	// Create the workflow
	metadata := map[string]string{
		"install_runbook_id":     installRunbook.ID,
		"install_runbook_run_id": run.ID,
		"runbook_name":           installRunbook.Runbook.Name,
		"runbook_config_id":      runbookConfig.ID,
	}

	workflow, err := s.installHelpers.CreateWorkflow(ctx, installID, app.WorkflowTypeRunbookRun, metadata, false)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create workflow: %w", err))
		return
	}

	// Link the run to the workflow
	s.db.WithContext(ctx).
		Model(&run).
		Update("install_workflow_id", workflow.ID)

	// Enqueue the workflow for execution
	queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install workflows queue: %w", err))
		return
	}

	if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
		WorkflowID: workflow.ID,
	}, workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue workflow: %w", err))
		return
	}

	run.InstallWorkflowID = &workflow.ID
	run.InstallWorkflow = workflow

	ctx.JSON(http.StatusCreated, run)
}

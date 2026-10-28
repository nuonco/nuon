package helpers

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/driftresolution"
)

func (h *Helpers) RecordInstallActionWorkflowRunApplied(ctx context.Context, runID string, appliedAt time.Time) error {
	var run app.InstallActionWorkflowRun
	if err := h.db.WithContext(ctx).
		Preload("ActionWorkflowConfig", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Where(app.InstallActionWorkflowRun{ID: runID}).
		First(&run).Error; err != nil {
		return fmt.Errorf("unable to get install action workflow run: %w", err)
	}
	if run.InstallActionWorkflowID.Empty() {
		return nil
	}
	if run.ActionWorkflowConfigID.Empty() {
		return nil
	}
	if run.ActionWorkflowConfig.AppConfigID == "" {
		return fmt.Errorf("action workflow config %s for run %s has no app config", run.ActionWorkflowConfigID.String, run.ID)
	}

	at := appliedAt
	if err := h.db.WithContext(ctx).
		Model(&app.InstallActionWorkflow{}).
		Where(app.InstallActionWorkflow{ID: run.InstallActionWorkflowID.String}).
		Update("app_config_ref", app.AppConfigRef{
			ExpectedConfigID:    run.ActionWorkflowConfig.AppConfigID,
			AppliedConfigID:     run.ActionWorkflowConfig.AppConfigID,
			AppliedConfigAt:     &at,
			AppliedConfigByType: app.AppConfigRefByTypeInstallActionWorkflowRuns,
			AppliedConfigByID:   run.ID,
		}).Error; err != nil {
		return fmt.Errorf("unable to record applied action workflow run: %w", err)
	}
	return nil
}

func (h *Helpers) RecordInstallStackVersionApplied(ctx context.Context, stackVersion *app.InstallStackVersion, requestType string) error {
	if requestType != PhoneHomeRequestTypeCreate && requestType != PhoneHomeRequestTypeUpdate {
		return nil
	}
	if stackVersion.InstallStackID == "" {
		return nil
	}

	appConfigID := stackVersion.AppConfigID

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := driftresolution.LockInstall(ctx, tx, stackVersion.InstallID); err != nil {
			return err
		}
		var existing app.InstallStack
		if err := tx.WithContext(ctx).
			Select("id", "app_config_ref").
			Where(app.InstallStack{ID: stackVersion.InstallStackID}).
			First(&existing).Error; err != nil {
			return fmt.Errorf("unable to get install stack: %w", err)
		}

		expectedConfigID := existing.AppConfigRef.ExpectedConfigID
		if appConfigID == "" {
			appConfigID = expectedConfigID
		}

		now := time.Now().UTC()
		if existing.AppConfigRef.AppliedConfigByID == stackVersion.ID && existing.AppConfigRef.AppliedConfigAt != nil {
			now = *existing.AppConfigRef.AppliedConfigAt
		}
		ref := app.AppConfigRef{
			ExpectedConfigID:    expectedConfigID,
			AppliedConfigID:     appConfigID,
			AppliedConfigAt:     &now,
			AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
			AppliedConfigByID:   stackVersion.ID,
		}

		if err := tx.WithContext(ctx).
			Model(&app.InstallStack{}).
			Where(app.InstallStack{ID: stackVersion.InstallStackID}).
			Update("app_config_ref", ref).Error; err != nil {
			return fmt.Errorf("unable to record actual stack version on install stack: %w", err)
		}
		return nil
	})
}

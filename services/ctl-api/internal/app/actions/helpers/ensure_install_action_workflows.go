package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	// batch sizes bound how much we hold in memory / send per statement so a large
	// app (many installs and/or many action workflows) can't OOM the api.
	actionWorkflowFetchBatchSize = 20
	installFetchBatchSize        = 20
	installActionInsertBatchSize = 20
)

func (h *Helpers) EnsureInstallAction(ctx context.Context, appID string, installIDs []string) error {
	// fan out over the app's action workflows in batches so we never load them all at once
	var actionWorkflows []app.ActionWorkflow
	return h.db.WithContext(ctx).
		Where(app.ActionWorkflow{AppID: appID}).
		FindInBatches(&actionWorkflows, actionWorkflowFetchBatchSize, func(_ *gorm.DB, _ int) error {
			// if install IDs are passed in, only ensure those
			if len(installIDs) > 0 {
				return h.createInstallActionWorkflows(ctx, installIDs, actionWorkflows)
			}

			// otherwise ensure every install for the app, also batched
			var installs []app.Install
			return h.db.WithContext(ctx).
				Where(app.Install{AppID: appID}).
				FindInBatches(&installs, installFetchBatchSize, func(_ *gorm.DB, _ int) error {
					batchIDs := make([]string, 0, len(installs))
					for _, install := range installs {
						batchIDs = append(batchIDs, install.ID)
					}
					return h.createInstallActionWorkflows(ctx, batchIDs, actionWorkflows)
				}).Error
		}).Error
}

func (h *Helpers) createInstallActionWorkflows(ctx context.Context, installIDs []string, actionWorkflows []app.ActionWorkflow) error {
	// build the cross product incrementally and flush at the batch size so we never
	// hold more than installActionInsertBatchSize insertable records in memory
	batch := make([]app.InstallActionWorkflow, 0, installActionInsertBatchSize)
	flush := func() error {
		if len(batch) < 1 {
			return nil
		}
		res := h.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&batch)
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to create install action workflows")
		}
		batch = batch[:0]
		return nil
	}

	for _, installID := range installIDs {
		for _, actionWorkflow := range actionWorkflows {
			batch = append(batch, app.InstallActionWorkflow{
				ActionWorkflowID: actionWorkflow.ID,
				InstallID:        installID,
			})
			if len(batch) < installActionInsertBatchSize {
				continue
			}
			if err := flush(); err != nil {
				return err
			}
		}
	}

	return flush()
}

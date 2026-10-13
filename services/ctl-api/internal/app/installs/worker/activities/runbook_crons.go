package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
)

type InstallRunbookCron struct {
	InstallRunbookID string `json:"install_runbook_id"`
	RunbookConfigID  string `json:"runbook_config_id"`
	TriggeredByID    string `json:"triggered_by_id"`
	CronSchedule     string `json:"cron_schedule"`
}

type GetInstallRunbookCronsRequest struct {
	InstallID   string `json:"install_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 1m
func (a *Activities) GetInstallRunbookCrons(ctx context.Context, req GetInstallRunbookCronsRequest) ([]InstallRunbookCron, error) {
	var configs []app.RunbookConfig
	if err := a.db.WithContext(ctx).
		Preload("Triggers").
		Where(app.RunbookConfig{AppConfigID: req.AppConfigID}).
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("get runbook configs: %w", err)
	}

	var installRunbooks []app.InstallRunbook
	if err := a.db.WithContext(ctx).
		Where(app.InstallRunbook{InstallID: req.InstallID}).
		Find(&installRunbooks).Error; err != nil {
		return nil, fmt.Errorf("get install runbooks: %w", err)
	}
	installRunbookIDs := make(map[string]string, len(installRunbooks))
	for _, installRunbook := range installRunbooks {
		installRunbookIDs[installRunbook.RunbookID] = installRunbook.ID
	}

	var result []InstallRunbookCron
	for _, config := range configs {
		installRunbookID := installRunbookIDs[config.RunbookID]
		if installRunbookID == "" {
			continue
		}
		for _, trigger := range config.Triggers {
			if trigger.Type != app.RunbookTriggerTypeCron {
				continue
			}
			result = append(result, InstallRunbookCron{
				InstallRunbookID: installRunbookID,
				RunbookConfigID:  config.ID,
				TriggeredByID:    config.CreatedByID,
				CronSchedule:     trigger.CronSchedule,
			})
		}
	}
	return result, nil
}

type TriggerRunbookCronRequest struct {
	InstallRunbookID string `json:"install_runbook_id" validate:"required"`
	RunbookConfigID  string `json:"runbook_config_id" validate:"required"`
	TriggeredByID    string `json:"triggered_by_id" validate:"required"`
	IdempotencyKey   string `json:"idempotency_key" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallRunbookID
// @start-to-close-timeout 1m
func (a *Activities) TriggerRunbookCron(ctx context.Context, req TriggerRunbookCronRequest) error {
	_, err := a.runbooksHelpers.TriggerRunbookRun(ctx, runbookshelpers.TriggerRunbookRunRequest{
		InstallRunbookID: req.InstallRunbookID,
		RunbookConfigID:  req.RunbookConfigID,
		TriggeredByID:    req.TriggeredByID,
		IdempotencyKey:   &req.IdempotencyKey,
	})
	return err
}

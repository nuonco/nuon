package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

type CreateProposedInstallsInput struct {
	AppID                  string                `json:"app_id" validate:"required"`
	AppInstallConfigSyncID string                `json:"app_install_config_sync_id" validate:"required_without=Names"`
	InstallConfigSyncID    string                `json:"install_config_sync_id" validate:"required_without=Names"`
	Installs               []app.ProposedInstall `json:"installs" validate:"required_without=Names"`
	Names                  []string              `json:"names,omitempty"`
}

type CreateProposedInstallsOutput struct {
	CreatedCount int `json:"created_count"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) CreateProposedInstalls(ctx context.Context, input *CreateProposedInstallsInput) (*CreateProposedInstallsOutput, error) {
	if len(input.Installs) == 0 && len(input.Names) > 0 {
		for _, name := range input.Names {
			if _, err := a.installHelpers.CreateInstall(ctx, input.AppID, &installhelpers.CreateInstallParams{Name: name}); err != nil {
				return nil, fmt.Errorf("unable to create install %s: %w", name, err)
			}
		}
		return &CreateProposedInstallsOutput{CreatedCount: len(input.Names)}, nil
	}

	record := app.InstallConfigSync{
		ID: input.InstallConfigSyncID, AppInstallConfigSyncID: &input.AppInstallConfigSyncID,
		TriggeredBy: "install-creation-approval", Status: app.NewCompositeStatus(ctx, app.StatusInProgress),
	}
	if err := a.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config sync: %w", err)
	}
	created := 0
	for _, proposal := range input.Installs {
		var installConfig config.Install
		if err := json.Unmarshal(proposal.Config, &installConfig); err != nil {
			return nil, fmt.Errorf("unable to decode proposed install %s: %w", proposal.Name, err)
		}
		if installConfig.Name != proposal.Name {
			return nil, fmt.Errorf("proposed install name does not match its config")
		}
		_, err := a.installActivities.SyncInstallConfig(ctx, &installactivities.SyncInstallConfigInput{
			AppID: input.AppID, InstallConfig: &installConfig, InstallConfigSyncID: record.ID,
			FilePath: proposal.FilePath, CreateOnly: true,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create install %s: %w", proposal.Name, err)
		}
		created++
	}
	if err := a.db.WithContext(ctx).Model(&app.InstallConfigSync{}).Where(app.InstallConfigSync{ID: record.ID}).
		Update("status", app.NewCompositeStatus(ctx, app.StatusSuccess)).Error; err != nil {
		return nil, err
	}

	return &CreateProposedInstallsOutput{CreatedCount: created}, nil
}

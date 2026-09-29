package helpers

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func (h *Helpers) MigrateInstallInputsToNewConfig(
	ctx context.Context,
	txn *gorm.DB,
	installConfigMap map[string]string,
	newAppConfigID string,
) error {
	if len(installConfigMap) == 0 {
		return nil
	}

	var newAppConfig app.AppConfig
	res := txn.WithContext(ctx).
		Where("id = ?", newAppConfigID).
		Preload("InputConfig").
		Preload("InputConfig.AppInputs").
		First(&newAppConfig)

	if res.Error != nil {
		return fmt.Errorf("unable to fetch new app config: %w", res.Error)
	}

	validInputs := make(map[string]bool)
	for _, inp := range newAppConfig.InputConfig.AppInputs {
		validInputs[inp.Name] = true
	}

	installIDs := make([]string, 0, len(installConfigMap))
	for installID := range installConfigMap {
		installIDs = append(installIDs, installID)
	}
	sort.Strings(installIDs)

	return txn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, installID := range installIDs {
			if err := LockInstallInputs(ctx, tx, installID); err != nil {
				return err
			}
			if err := h.migrateInstallInputs(ctx, tx, installID, newAppConfig.InputConfig.ID, validInputs); err != nil {
				return fmt.Errorf("unable to migrate inputs for install %s: %w", installID, err)
			}

			if err := h.MarkInstallStatePartialsStale(ctx, tx, installID,
				state.HintToPartials[state.HintAppConfigUpdated]...); err != nil {
				return fmt.Errorf("unable to mark state partials stale for install %s: %w", installID, err)
			}
		}
		return nil
	})
}

func (h *Helpers) migrateInstallInputs(
	ctx context.Context,
	txn *gorm.DB,
	installID string,
	newAppInputConfigID string,
	validInputs map[string]bool,
) error {
	var onNewConfig int64
	if err := txn.WithContext(ctx).
		Model(&app.InstallInputs{}).
		Where(app.InstallInputs{
			InstallID:        installID,
			AppInputConfigID: newAppInputConfigID,
		}).
		Count(&onNewConfig).Error; err != nil {
		return fmt.Errorf("unable to check inputs on new config: %w", err)
	}

	if onNewConfig == 0 {
		// why: newest row, not the outgoing config's: that read cannot see a write pinned to the new one
		var existingInputs app.InstallInputs
		res := txn.WithContext(ctx).
			Where(app.InstallInputs{
				InstallID: installID,
			}).
			Order("created_at DESC").
			Limit(1).
			Find(&existingInputs)
		if res.Error != nil {
			return errors.Wrap(res.Error, fmt.Sprintf("unable to fetch install inputs for installID %s", installID))
		}

		migratedValues := make(pgtype.Hstore)
		for key, value := range existingInputs.Values {
			if validInputs[key] {
				migratedValues[key] = value
			}
		}

		if len(migratedValues) != len(existingInputs.Values) {
			newInputs := app.InstallInputs{
				InstallID:        installID,
				AppInputConfigID: newAppInputConfigID,
				Values:           migratedValues,
			}
			if err := txn.WithContext(ctx).Create(&newInputs).Error; err != nil {
				return fmt.Errorf("unable to create migrated inputs: %w", err)
			}
		}
	}

	return nil
}

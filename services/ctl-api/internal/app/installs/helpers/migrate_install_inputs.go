package helpers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// MigrateInstallInputsToNewConfig creates new InstallInputs records for installs
// when their app_config_id is updated. Preserves existing values where field names
// match, drops fields removed in new config.
func (h *Helpers) MigrateInstallInputsToNewConfig(
	ctx context.Context,
	txn *gorm.DB,
	installConfigMap map[string]string, // installID -> old appConfigID
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

	for installID, oldAppConfigID := range installConfigMap {
		if err := h.migrateInstallInputs(
			ctx,
			txn,
			installID,
			oldAppConfigID,
			newAppConfig.InputConfig.ID,
			validInputs); err != nil {
			return fmt.Errorf("unable to migrate inputs for install %s: %w", installID, err)
		}
	}

	return nil
}

func (h *Helpers) migrateInstallInputs(
	ctx context.Context,
	txn *gorm.DB,
	installID string,
	oldAppConfigID string,
	newAppInputConfigID string,
	validInputs map[string]bool,
) error {
	var oldAppConfig app.AppConfig
	res := txn.WithContext(ctx).
		Where("id = ?", oldAppConfigID).
		Preload("InputConfig").
		First(&oldAppConfig)

	if res.Error != nil {
		return fmt.Errorf("unable to fetch old app config: %w", res.Error)
	}

	var existingInputs app.InstallInputs
	res = txn.WithContext(ctx).
		Where(app.InstallInputs{
			InstallID:        installID,
			AppInputConfigID: oldAppConfig.InputConfig.ID,
		}).
		Order("created_at DESC").
		Limit(1).
		Find(&existingInputs)

	if res.Error != nil && res.Error == gorm.ErrRecordNotFound {
		// for backward compatibility for older installs where older installs dont have install in puts set
		// at latest app config
		res := txn.WithContext(ctx).
			Where(app.InstallInputs{
				InstallID: installID,
			}).
			Order("created_at DESC").
			Limit(1).
			Find(&existingInputs)

		// error out if install exists but there are no install inputs associated with it
		if res.Error != nil {
			return errors.Wrap(res.Error, fmt.Sprintf("unable to fetch install inputs for installID %s", installID))
		}
	} else if res.Error != nil {
		return fmt.Errorf("unable to fetch existing inputs: %w", res.Error)
	}

	migratedValues := make(pgtype.Hstore)
	for key, value := range existingInputs.Values {
		if validInputs[key] {
			migratedValues[key] = value
		}
	}

	newInputs := app.InstallInputs{
		InstallID:        installID,
		AppInputConfigID: newAppInputConfigID,
		Values:           migratedValues,
		OrgID:            existingInputs.OrgID,
	}

	if err := txn.WithContext(ctx).Create(&newInputs).Error; err != nil {
		return fmt.Errorf("unable to create migrated inputs: %w", err)
	}

	return nil
}

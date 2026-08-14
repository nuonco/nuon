package helpers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (h *Helpers) UpdateInstallInputsFromStackOutputs(
	ctx context.Context,
	installStackVersionID,
	installID,
	inputConfigID string,
	inputValues map[string]string,
	skipInputUpdateWorkflow bool,
) (*app.Workflow, error) {
	if len(inputValues) == 0 {
		return nil, nil
	}
	install, err := h.getInstall(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install: "+installID)
	}

	var account app.Account
	res := h.db.WithContext(ctx).
		Where(app.Account{
			Subject:     installStackVersionID,
			AccountType: app.AccountTypeService,
		}).
		First(&account)
	if res.Error != nil {
		return nil, errors.Wrap(
			res.Error,
			"unable to fetch service account for install stack version: "+installStackVersionID,
		)
	}

	ctx = cctx.SetAccountIDContext(ctx, account.ID)
	ctx = cctx.SetOrgIDContext(ctx, install.OrgID)

	var appInputConfig app.AppInputConfig
	if res := h.db.WithContext(ctx).
		Preload("AppInputs").
		Where("id = ?", inputConfigID).
		First(&appInputConfig); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app input config")
	}

	validInputs := make(map[string]bool)
	for _, input := range appInputConfig.AppInputs {
		if input.Source == app.AppInputSourceCustomer {
			validInputs[input.Name] = true
		}
	}

	filteredInputValues := make(map[string]string)
	for key, value := range inputValues {
		if validInputs[key] {
			filteredInputValues[key] = value
		}
	}

	if len(filteredInputValues) == 0 {
		return nil, nil
	}

	newValuesPtr := make(map[string]*string)
	for k, v := range filteredInputValues {
		newValuesPtr[k] = generics.ToPtr(v)
	}

	var changed *ChangedInputsResult
	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := LockInstallInputs(ctx, tx, installID); err != nil {
			return err
		}

		// newest row for the install: a row this config no longer pins is never read
		var installInputs app.InstallInputs
		res := tx.WithContext(ctx).
			Where(app.InstallInputs{
				InstallID: installID,
			}).
			Order("created_at DESC").
			Limit(1).
			Find(&installInputs)
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to get install inputs")
		}

		if installInputs.Values == nil {
			installInputs.Values = pgtype.Hstore{}
		}

		var err error
		changed, err = ComputeChangedInputs(
			installInputs.Values,
			newValuesPtr,
			appInputConfig.AppInputs,
		)
		if err != nil {
			return errors.Wrap(err, "unable to compute changed inputs")
		}

		for key, value := range filteredInputValues {
			installInputs.Values[key] = generics.ToPtr(value)
		}

		if res.RowsAffected == 0 {
			newInputs := app.InstallInputs{
				InstallID:        installID,
				AppInputConfigID: appInputConfig.ID,
				Values:           installInputs.Values,
			}
			if err := tx.WithContext(ctx).Create(&newInputs).Error; err != nil {
				return errors.Wrap(err, "unable to create install inputs")
			}
			return nil
		}

		if err := tx.WithContext(ctx).
			Model(&app.InstallInputs{}).
			Where("id = ?", installInputs.ID).
			Update("values", installInputs.Values).Error; err != nil {
			return errors.Wrap(err, "unable to update install inputs")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(changed.Names) > 0 && !skipInputUpdateWorkflow {
		wkflw, err := h.CreateAndStartInputUpdateWorkflow(
			ctx,
			installID,
			changed.Names,
			changed.ChangedValuesJSON,
			"",
			true,
			false,
			false,
			app.WorkflowTypeInputUpdate,
		)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update inputs from install stack output")
		}
		return wkflw, nil
	}

	return nil, nil
}

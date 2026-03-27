package helpers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// UpdateInstallInputsFromStackOutputs updates install_inputs table with values from stack outputs
// inputValues should only contain the app inputs extracted from stack outputs (from the "app_inputs" nested object)
func (h *Helpers) UpdateInstallInputsFromStackOutputs(
	ctx context.Context,
	installStackVersionID,
	installID,
	inputConfigID string,
	inputValues map[string]string,
	skipInputUpdateWorkflow bool,
) error {
	// If no input values provided, nothing to update
	if len(inputValues) == 0 {
		return nil
	}
	install, err := h.getInstall(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install: "+installID)
	}

	// update contexe with org and service account information
	var account app.Account
	res := h.db.WithContext(ctx).
		Where(app.Account{
			Subject:     installStackVersionID,
			AccountType: app.AccountTypeService,
		}).
		First(&account)
	if res.Error != nil {
		return errors.Wrap(
			res.Error,
			"unable to fetch service account for install stack version: "+installStackVersionID,
		)
	}

	ctx = cctx.SetAccountIDContext(ctx, account.ID)
	ctx = cctx.SetOrgIDContext(ctx, install.OrgID)

	// Get the app input config to validate inputs
	var appInputConfig app.AppInputConfig
	if res := h.db.WithContext(ctx).
		Preload("AppInputs").
		Where("id = ?", inputConfigID).
		First(&appInputConfig); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get app input config")
	}

	// Validate that inputs are actually install_stack sourced
	validInputs := make(map[string]bool)
	for _, input := range appInputConfig.AppInputs {
		if input.Source == app.AppInputSourceCustomer {
			validInputs[input.Name] = true
		}
	}

	// Filter input values to only include valid install_stack sourced inputs
	filteredInputValues := make(map[string]string)
	for key, value := range inputValues {
		if validInputs[key] {
			filteredInputValues[key] = value
		}
	}

	if len(filteredInputValues) == 0 {
		return nil
	}

	// Get or create install inputs record
	var installInputs app.InstallInputs
	res = h.db.WithContext(ctx).
		Where(app.InstallInputs{
			InstallID:        installID,
			AppInputConfigID: appInputConfig.ID,
		}).
		Attrs(app.InstallInputs{Values: pgtype.Hstore{}}).
		FirstOrCreate(&installInputs)

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to get or create install inputs")
	}

	// Merge new values with existing values and track changes
	if installInputs.Values == nil {
		return errors.New("missing install inputs")
	}

	// Convert filtered values to pointer map for comparison
	newValuesPtr := make(map[string]*string)
	for k, v := range filteredInputValues {
		newValuesPtr[k] = generics.ToPtr(v)
	}

	changed, err := ComputeChangedInputs(
		installInputs.Values,
		newValuesPtr,
		appInputConfig.AppInputs,
	)
	if err != nil {
		return errors.Wrap(err, "unable to compute changed inputs")
	}

	// Merge new values into existing map
	newInputMap := installInputs.Values
	for key, value := range filteredInputValues {
		newInputMap[key] = generics.ToPtr(value)
	}

	// Update the install inputs with merged values
	installInputs.Values = newInputMap

	if res := h.db.WithContext(ctx).
		Model(&installInputs).
		Where("id = ?", installInputs.ID).
		Update("values", installInputs.Values); res.Error != nil {
		return errors.Wrap(res.Error, "unable to update install inputs")
	}

	if len(changed.Names) > 0 && !skipInputUpdateWorkflow {
		// Send signals to notify that inputs have been updated from stack outputs
		_, err = h.CreateAndStartInputUpdateWorkflow(
			ctx,
			installID,
			changed.Names,
			changed.ChangedValuesJSON,
			"",
			true,
		)
		if err != nil {
			return errors.Wrap(err, "unable to update inputs from install stack output")
		}

		// Send signals to notify that inputs have been updated from stack outputs.
		// NOTE: passes nil v2Signals because this is called from a Temporal activity (worker/activities),
		// and importing v2 signal packages would create an import cycle (helpers → v2 → worker → activities → helpers).
		// Legacy signals are used as fallback.
		//_, err = h.CreateAndStartInputUpdateWorkflow(ctx, installID, changed.ChangedValuesJSON)
		//if err != nil {
		//return errors.Wrap(err, "unable to update inputs from install stack output")
		//}
	}

	return nil
}

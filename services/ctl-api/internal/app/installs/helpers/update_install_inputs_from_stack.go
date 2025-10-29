package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// UpdateInstallInputsFromStackOutputs updates install_inputs table with values from stack outputs
// inputValues should only contain the app inputs extracted from stack outputs (from the "app_inputs" nested object)
func (h *Helpers) UpdateInstallInputsFromStackOutputs(ctx context.Context, installStackVersionID, installID, inputConfigID string, inputValues map[string]string) error {
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
		FirstOrCreate(&installInputs)

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to get or create install inputs")
	}

	// Merge new values with existing values and track changes
	newInputMap := installInputs.Values
	var changedInputs []string
	for key, value := range filteredInputValues {
		// Check if value is different from existing
		if newInputMap[key] == nil || *newInputMap[key] != value {
			changedInputs = append(changedInputs, key)
		}
		newInputMap[key] = &value
	}

	// Update the install inputs with merged values
	installInputs.Values = newInputMap

	if res := h.db.WithContext(ctx).
		Model(&installInputs).
		Where("id = ?", installInputs.ID).
		Update("values", installInputs.Values); res.Error != nil {
		return errors.Wrap(res.Error, "unable to update install inputs")
	}

	// Send signals to notify that inputs have been updated from stack outputs
	_, err = h.CreateAndStartInputUpdateWorkflow(ctx, installID, changedInputs)
	if err != nil {
		return errors.Wrap(err, "unable to update inputs from install stack output")
	}

	return nil
}

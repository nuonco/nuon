package helpers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

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

	var installInputs app.InstallInputs
	res = h.db.WithContext(ctx).
		Where(app.InstallInputs{
			InstallID:        installID,
			AppInputConfigID: appInputConfig.ID,
		}).
		Order("created_at DESC").
		Attrs(app.InstallInputs{Values: pgtype.Hstore{}}).
		FirstOrCreate(&installInputs)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get or create install inputs")
	}

	if installInputs.Values == nil {
		installInputs.Values = pgtype.Hstore{}
	}

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
		return nil, errors.Wrap(err, "unable to compute changed inputs")
	}

	newInputMap := installInputs.Values
	for key, value := range filteredInputValues {
		newInputMap[key] = generics.ToPtr(value)
	}

	installInputs.Values = newInputMap

	if res := h.db.WithContext(ctx).
		Model(&installInputs).
		Where("id = ?", installInputs.ID).
		Update("values", installInputs.Values); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update install inputs")
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
			app.WorkflowTypeInputUpdate,
		)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update inputs from install stack output")
		}
		return wkflw, nil
	}

	return nil, nil
}

package helpers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ApplyInputPatchResult captures the output of applying a partial input update.
type ApplyInputPatchResult struct {
	// InstallInputs is the freshly persisted InstallInputs row (Values cleared,
	// matching the HTTP handler's response shape).
	InstallInputs *app.InstallInputs
	// ChangedNames is the list of input names whose values changed.
	ChangedNames []string
	// ChangedValuesJSON is a JSON snapshot of old→new values for audit logging
	// (sensitive inputs are masked).
	ChangedValuesJSON string
}

// ApplyInstallInputPatch merges a partial input patch over the install's
// current inputs, validates the merged set against the pinned app input
// config, persists a new InstallInputs row, and returns the changed input
// names + audit values.
//
// Shared by the PATCH /v1/installs/{id}/inputs HTTP handler and the runbook
// input_update step. Callers are responsible for starting any workflow that
// reacts to the change.
func (h *Helpers) ApplyInstallInputPatch(
	ctx context.Context,
	install *app.Install,
	patch map[string]*string,
) (*ApplyInputPatchResult, error) {
	if len(install.App.AppInputConfigs) < 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("no app input configs defined on app"),
			Description: "no app input configs defined",
		}
	}

	latestInstallInputs := app.InstallInputs{}
	if err := h.db.WithContext(ctx).
		Where("install_id = ?", install.ID).
		Order("created_at DESC").
		First(&latestInstallInputs).Error; err != nil {
		return nil, fmt.Errorf("unable to get latest install inputs: %w", err)
	}

	pinnedAppInputConfig, err := h.GetPinnedAppInputConfig(ctx, install.AppID, install.AppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to get pinned app input config: %w", err)
	}
	if pinnedAppInputConfig == nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("invalid install inputs provided"),
			Description: "inputs provided on install, that are not defined on the app",
		}
	}

	if err := validateVendorSourceInputs(pinnedAppInputConfig, patch); err != nil {
		return nil, err
	}

	merged := MergeInstallInputs(latestInstallInputs.Values, patch, pinnedAppInputConfig)
	if err := h.ValidateInstallInputs(ctx, pinnedAppInputConfig, merged); err != nil {
		return nil, err
	}

	changed, err := ComputeChangedInputs(
		latestInstallInputs.Values,
		patch,
		pinnedAppInputConfig.AppInputs,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to compute changed inputs: %w", err)
	}

	obj := &app.InstallInputs{
		AppInputConfigID: pinnedAppInputConfig.ID,
		InstallID:        install.ID,
		Values:           pgtype.Hstore(merged),
	}
	if err := h.db.WithContext(ctx).Create(&obj).Error; err != nil {
		return nil, fmt.Errorf("unable to create install inputs: %w", err)
	}

	// Re-fetch so we return the latest row with its generated fields, then clear
	// Values for the response (matching the HTTP handler's contract).
	persisted := app.InstallInputs{}
	if err := h.db.WithContext(ctx).
		Where("install_id = ?", install.ID).
		Order("created_at DESC").
		First(&persisted).Error; err != nil {
		return nil, fmt.Errorf("unable to get latest install inputs after create: %w", err)
	}
	persisted.Values = nil

	return &ApplyInputPatchResult{
		InstallInputs:     &persisted,
		ChangedNames:      changed.Names,
		ChangedValuesJSON: changed.ChangedValuesJSON,
	}, nil
}

// MergeInstallInputs overlays the provided subset onto the install's existing
// input values and drops any inputs no longer defined in the pinned app input
// config.
func MergeInstallInputs(existing map[string]*string, patch map[string]*string, appInputConfig *app.AppInputConfig) map[string]*string {
	merged := map[string]*string{}
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range patch {
		merged[k] = v
	}

	appInputNames := map[string]struct{}{}
	for _, input := range appInputConfig.AppInputs {
		appInputNames[input.Name] = struct{}{}
	}
	for k := range merged {
		if _, ok := appInputNames[k]; !ok {
			delete(merged, k)
		}
	}

	return merged
}

func validateVendorSourceInputs(appInputConfig *app.AppInputConfig, inputs map[string]*string) error {
	appInputSources := map[string]app.AppInputSource{}
	for _, input := range appInputConfig.AppInputs {
		appInputSources[input.Name] = input.Source
	}

	for name := range inputs {
		source, ok := appInputSources[name]
		if !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("input %s is not defined in app input config", name),
				Description: "input " + name + " does not exist in the app inputs",
			}
		}

		if source == app.AppInputSourceCustomer {
			return stderr.ErrUser{
				Err:         fmt.Errorf("%s has source install_stack, cannot be updated via api", name),
				Description: name + " has source install_stack and cannot be updated via the api",
			}
		}
	}

	return nil
}

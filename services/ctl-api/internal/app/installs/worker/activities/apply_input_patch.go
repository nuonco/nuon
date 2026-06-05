package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ApplyInstallInputPatchRequest struct {
	InstallID string             `validate:"required"`
	Patch     map[string]*string `validate:"required"`
}

type ApplyInstallInputPatchResponse struct {
	InstallInputsID   string   `json:"install_inputs_id"`
	ChangedNames      []string `json:"changed_names"`
	ChangedValuesJSON string   `json:"changed_values_json"`
}

// ApplyInstallInputPatch merges the patch over the install's current inputs,
// validates the result, and persists a new InstallInputs row. Returns the list
// of input names whose values changed so the caller can drive the input-update
// step sequence.
//
// @temporal-gen-v2 activity
func (a *Activities) ApplyInstallInputPatch(ctx context.Context, req *ApplyInstallInputPatchRequest) (*ApplyInstallInputPatchResponse, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).
		Preload("App.AppInputConfigs").
		Where("id = ?", req.InstallID).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to load install %s: %w", req.InstallID, err)
	}

	result, err := a.helpers.ApplyInstallInputPatch(ctx, &install, req.Patch)
	if err != nil {
		return nil, err
	}

	return &ApplyInstallInputPatchResponse{
		InstallInputsID:   result.InstallInputs.ID,
		ChangedNames:      result.ChangedNames,
		ChangedValuesJSON: result.ChangedValuesJSON,
	}, nil
}

package activities

import (
	"context"
)

type UpdateInstallInputsFromStackRequest struct {
	InstallID               string            `temporaljson:"install_id"`
	InputConfigID           string            `temporaljson:"input_config_id"`
	InputValues             map[string]string `temporaljson:"input_values"`
	InstallStackVersionID   string            `temporaljson:"install_stack_version_id"`
	SkipInputUpdateWorkflow bool              `temporaljson:"skip_input_update_workflow"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateInstallInputsFromStack(ctx context.Context, req *UpdateInstallInputsFromStackRequest) error {
	return a.helpers.UpdateInstallInputsFromStackOutputs(
		ctx,
		req.InstallStackVersionID,
		req.InstallID,
		req.InputConfigID,
		req.InputValues,
		req.SkipInputUpdateWorkflow,
	)
}

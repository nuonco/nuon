package activities

import (
	"context"
	"fmt"
)

type IsRunnerDisabledRequest struct {
	InstallID string `validate:"required"`
}

type IsRunnerDisabledResponse struct {
	Disabled bool `json:"disabled"`
}

// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) IsRunnerDisabled(ctx context.Context, req IsRunnerDisabledRequest) (*IsRunnerDisabledResponse, error) {
	disabled, err := a.helpers.IsRunnerDisabled(ctx, req.InstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to determine whether the install runner is disabled: %w", err)
	}

	return &IsRunnerDisabledResponse{Disabled: disabled}, nil
}

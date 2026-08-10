package activities

import (
	"context"
	"fmt"

	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

type CreateProposedInstallsInput struct {
	AppID string   `json:"app_id" validate:"required"`
	Names []string `json:"names" validate:"required"`
}

type CreateProposedInstallsOutput struct {
	CreatedCount int `json:"created_count"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) CreateProposedInstalls(ctx context.Context, input *CreateProposedInstallsInput) (*CreateProposedInstallsOutput, error) {
	created := 0
	for _, name := range input.Names {
		_, err := a.installHelpers.CreateInstall(ctx, input.AppID, &installshelpers.CreateInstallParams{
			Name: name,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create install %s: %w", name, err)
		}
		created++
	}

	return &CreateProposedInstallsOutput{CreatedCount: created}, nil
}

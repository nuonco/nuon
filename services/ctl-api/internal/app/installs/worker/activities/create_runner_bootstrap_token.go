package activities

import (
	"context"

	"github.com/pkg/errors"
)

type CreateRunnerBootstrapTokenRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) CreateRunnerBootstrapTokenRequest(ctx context.Context, req *CreateRunnerBootstrapTokenRequest) (*string, error) {
	token, err := a.runnersHelpers.CreateBootstrapToken(ctx, req.RunnerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create bootstrap token")
	}

	return &token.Token, nil
}

package activities

import (
	"context"

	"github.com/pkg/errors"
)

// NOTE(JM): this is a short term thing until we are doing cross namespace steps
type CreateRunnerTokenRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen activity
// @by-id RunnerID
func (a *Activities) CreateRunnerTokenRequest(ctx context.Context, req *CreateRunnerTokenRequest) (*string, error) {
	token, err := a.runnersHelpers.CreateToken(ctx, req.RunnerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	return &token.Token, nil
}

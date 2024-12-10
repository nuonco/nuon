package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

const (
	defaultRunnerTokenTimeout time.Duration = time.Hour * 24 * 90
)

type CreateTokenRequest struct {
	RunnerID string `validate:"required"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}

// @temporal-gen activity
// @by-id RunnerID
func (a *Activities) CreateJobLogToken(ctx context.Context, req CreateTokenRequest) (*CreateTokenResponse, error) {
	email := account.ServiceAccountEmail(req.RunnerID)

	token, err := a.acctClient.CreateToken(ctx, email, defaultRunnerTokenTimeout)
	if err != nil {
		return nil, fmt.Errorf("unable to create token: %w", err)
	}

	return &CreateTokenResponse{
		Token: token.Token,
	}, nil
}

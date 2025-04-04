package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

const (
	defaultActionWorkflowRunTimeout time.Duration = time.Minute * 60
)

type CreateActionWorkflowRunTokenRequest struct {
	RunnerID string `validate:"required"`
}

type CreateActionWorkflowRunTokenResponse struct {
	Token  string `json:"token"`
	APIURL string `json:"api_url"`
}

// @temporal-gen activity
// @by-id RunnerID
func (a *Activities) CreateActionWorkflowRunToken(ctx context.Context, req *CreateActionWorkflowRunTokenRequest) (*CreateActionWorkflowRunTokenResponse, error) {
	email := account.ServiceAccountEmail(req.RunnerID)

	token, err := a.acctClient.CreateToken(ctx, email, defaultActionWorkflowRunTimeout)
	if err != nil {
		return nil, fmt.Errorf("unable to create token: %w", err)
	}

	return &CreateActionWorkflowRunTokenResponse{
		Token:  token.Token,
		APIURL: a.cfg.PublicAPIURL,
	}, nil
}

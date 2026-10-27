package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	defaultRunnerTokenTimeout time.Duration = time.Hour * 24 * 90
)

type CreateTokenRequest struct {
	RunnerID string `validate:"required"`
	JobID    string
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) CreateJobLogToken(ctx context.Context, req CreateTokenRequest) (*CreateTokenResponse, error) {
	email := account.ServiceAccountEmail(req.RunnerID)
	dur := defaultRunnerTokenTimeout
	issue := account.IssueTokenRequest{}
	if req.JobID != "" {
		issue = account.IssueTokenRequest{SourceType: app.TokenSourceTypeRunnerJob, SourceID: req.JobID}
		jobEmail := account.ServiceAccountEmail(req.JobID)
		if _, err := a.acctClient.FindAccount(ctx, jobEmail); err == nil {
			email = jobEmail
			dur = time.Hour * 6
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("lookup job service account: %w", err)
		}
	}

	token, err := a.acctClient.IssueToken(ctx, email, dur, issue)
	if err != nil {
		return nil, fmt.Errorf("unable to create token: %w", err)
	}

	return &CreateTokenResponse{
		Token: token.Token,
	}, nil
}

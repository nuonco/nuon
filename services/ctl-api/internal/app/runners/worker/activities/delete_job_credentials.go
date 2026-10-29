package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

type DeleteJobCredentialsRequest struct {
	JobID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @max-retries 3
func (a *Activities) DeleteJobCredentials(ctx context.Context, req DeleteJobCredentialsRequest) error {
	if err := account.RevokeDerivedCredentials(a.db.WithContext(ctx), app.TokenSourceTypeRunnerJob, []string{req.JobID}); err != nil {
		return fmt.Errorf("revoke job tokens: %w", err)
	}
	if err := a.acctClient.DeleteServiceAccount(ctx, req.JobID); err != nil {
		return fmt.Errorf("delete job service account: %w", err)
	}
	return nil
}

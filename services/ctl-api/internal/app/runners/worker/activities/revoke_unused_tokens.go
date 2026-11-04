package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
)

type RevokeUnusedTokensRequest struct {
	RunnerID string `validate:"required" json:"runner_id" temporaljson:"runner_id,omitzero,omitempty"`
}

type RevokeUnusedTokensResponse struct {
	RevokedCount int64 `json:"revoked_count" temporaljson:"revoked_count,omitzero,omitempty"`
}

// @temporal-gen-v2 activity
func (a *Activities) RevokeUnusedTokens(ctx context.Context, req RevokeUnusedTokensRequest) (*RevokeUnusedTokensResponse, error) {
	count, err := a.helpers.RevokeUnusedTokens(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, helpers.ErrTokenCleanupNotReady) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "token_cleanup_not_ready", err)
		}
		return nil, errors.Wrap(err, "unable to revoke unused tokens")
	}

	return &RevokeUnusedTokensResponse{RevokedCount: count}, nil
}

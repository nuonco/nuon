package activities

import (
	"context"

	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "unable to revoke unused tokens")
	}

	return &RevokeUnusedTokensResponse{RevokedCount: count}, nil
}

package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type DeleteOldOIDCExchangedTokensRequest struct {
	BatchSize int `json:"batch_size"`
}

type DeleteOldOIDCExchangedTokensResponse struct {
	Deleted int64 `json:"deleted"`
}

// DeleteOldOIDCExchangedTokens hard-deletes a single batch of OIDCExchangedToken
// rows whose source JWT has already expired. ExpiresAt is the source token's own
// exp claim, so once it has passed the token can never verify again and the
// replay-protection row no longer needs to be checked.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) DeleteOldOIDCExchangedTokens(ctx context.Context, req DeleteOldOIDCExchangedTokensRequest) (*DeleteOldOIDCExchangedTokensResponse, error) {
	if req.BatchSize <= 0 {
		req.BatchSize = 5000
	}

	sub := a.db.Model(&app.OIDCExchangedToken{}).
		Unscoped().
		Select("id").
		Where("expires_at < ?", time.Now()).
		Limit(req.BatchSize)

	res := a.db.WithContext(ctx).
		Unscoped().
		Where("id IN (?)", sub).
		Delete(&app.OIDCExchangedToken{})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to delete old oidc exchanged tokens: %w", res.Error)
	}

	a.mw.Count("general.oidc_exchanged_token_cleanup.deleted", res.RowsAffected, []string{})
	return &DeleteOldOIDCExchangedTokensResponse{Deleted: res.RowsAffected}, nil
}

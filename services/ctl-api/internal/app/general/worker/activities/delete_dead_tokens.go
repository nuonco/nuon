package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	revokedTokenRetention = 24 * time.Hour
	expiredTokenRetention = 30 * 24 * time.Hour
)

type DeleteDeadTokensRequest struct {
	AfterID   string `json:"after_id"`
	BatchSize int    `json:"batch_size"`
}

type DeleteDeadTokensResponse struct {
	account.DeleteDeadTokensResult
	Enabled bool `json:"enabled"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) DeleteDeadTokens(ctx context.Context, req DeleteDeadTokensRequest) (*DeleteDeadTokensResponse, error) {
	mode := a.cfg.DeadTokenCleanupMode
	if mode != "on" && mode != "dry-run" {
		return &DeleteDeadTokensResponse{}, nil
	}

	now := time.Now()
	res, err := a.acctClient.DeleteDeadTokens(ctx, account.DeleteDeadTokensRequest{
		AfterID:       req.AfterID,
		Limit:         req.BatchSize,
		RevokedBefore: now.Add(-revokedTokenRetention),
		ExpiredBefore: now.Add(-expiredTokenRetention),
		DryRun:        mode == "dry-run",
	})
	if err != nil {
		return nil, fmt.Errorf("unable to delete dead tokens: %w", err)
	}

	tags := metrics.ToTags(map[string]string{"mode": mode})
	a.mw.Count("general.dead_token_cleanup.selected", int64(res.Selected), tags)
	a.mw.Count("general.dead_token_cleanup.deleted", int64(res.Deleted), tags)

	return &DeleteDeadTokensResponse{DeleteDeadTokensResult: *res, Enabled: true}, nil
}

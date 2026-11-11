package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	logStreamAccountIdleGrace = 2 * time.Hour
	logStreamOpenMaxAge       = 24 * time.Hour
)

type SweepLogStreamAccountsRequest struct {
	AfterID   string `json:"after_id"`
	BatchSize int    `json:"batch_size"`
}

type SweepLogStreamAccountsResponse struct {
	account.SweepLogStreamAccountsResult
	Enabled bool `json:"enabled"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) SweepLogStreamAccounts(ctx context.Context, req SweepLogStreamAccountsRequest) (*SweepLogStreamAccountsResponse, error) {
	mode := a.cfg.LogStreamAccountSweepMode
	if mode != "on" && mode != "dry-run" {
		return &SweepLogStreamAccountsResponse{}, nil
	}

	now := time.Now()
	res, err := a.acctClient.SweepLogStreamAccounts(ctx, account.SweepLogStreamAccountsRequest{
		AfterID:          req.AfterID,
		Limit:            req.BatchSize,
		IdleBefore:       now.Add(-logStreamAccountIdleGrace),
		OpenStreamBefore: now.Add(-logStreamOpenMaxAge),
		DryRun:           mode == "dry-run",
	})
	if err != nil {
		return nil, fmt.Errorf("unable to sweep log stream accounts: %w", err)
	}

	tags := metrics.ToTags(map[string]string{"mode": mode})
	a.mw.Count("general.log_stream_account_sweep.selected", int64(res.Selected), tags)
	a.mw.Count("general.log_stream_account_sweep.deleted", int64(res.Deleted), tags)
	a.mw.Count("general.log_stream_account_sweep.skipped", int64(res.Skipped), tags)
	a.mw.Count("general.log_stream_account_sweep.streams_closed", int64(res.StreamsClosed), tags)

	return &SweepLogStreamAccountsResponse{SweepLogStreamAccountsResult: *res, Enabled: true}, nil
}

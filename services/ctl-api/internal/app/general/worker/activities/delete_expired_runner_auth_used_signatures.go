package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type DeleteExpiredRunnerAuthUsedSignaturesRequest struct {
	BatchSize int `json:"batch_size"`
}

type DeleteExpiredRunnerAuthUsedSignaturesResponse struct {
	Deleted int64 `json:"deleted"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) DeleteExpiredRunnerAuthUsedSignatures(ctx context.Context, req DeleteExpiredRunnerAuthUsedSignaturesRequest) (*DeleteExpiredRunnerAuthUsedSignaturesResponse, error) {
	if req.BatchSize <= 0 {
		req.BatchSize = 5000
	}

	sub := a.db.Model(&app.RunnerAuthUsedSignature{}).
		Unscoped().
		Select("id").
		Where("expires_at < ?", time.Now()).
		Limit(req.BatchSize)

	res := a.db.WithContext(ctx).
		Unscoped().
		Where("id IN (?)", sub).
		Delete(&app.RunnerAuthUsedSignature{})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to delete expired runner auth used signatures: %w", res.Error)
	}

	a.mw.Count("general.runner_auth_used_signature_cleanup.deleted", res.RowsAffected, []string{})
	return &DeleteExpiredRunnerAuthUsedSignaturesResponse{Deleted: res.RowsAffected}, nil
}

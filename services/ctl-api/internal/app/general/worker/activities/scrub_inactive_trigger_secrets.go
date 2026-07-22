package activities

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ScrubInactiveTriggerSecretsRequest struct {
	BatchSize int `json:"batch_size"`
}

type ScrubInactiveTriggerSecretsResponse struct {
	Scrubbed int64 `json:"scrubbed"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) ScrubInactiveTriggerSecrets(ctx context.Context, req ScrubInactiveTriggerSecretsRequest) (*ScrubInactiveTriggerSecretsResponse, error) {
	if req.BatchSize <= 0 {
		req.BatchSize = 5000
	}
	now := time.Now()
	var scrubbed int64
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var candidates []app.TriggerSecret
		if err := tx.Select("id").
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where(clause.Neq{Column: "secret", Value: ""}).
			Where(clause.Or(clause.Neq{Column: "revoked_at", Value: nil}, clause.Lte{Column: "expires_at", Value: now})).
			Order("id").Limit(req.BatchSize).Find(&candidates).Error; err != nil {
			return err
		}
		if len(candidates) == 0 {
			return nil
		}
		ids := make([]string, len(candidates))
		for i := range candidates {
			ids[i] = candidates[i].ID
		}
		res := tx.Model(&app.TriggerSecret{}).
			Where("id IN ?", ids).
			Where(clause.Neq{Column: "secret", Value: ""}).
			Where(clause.Or(clause.Neq{Column: "revoked_at", Value: nil}, clause.Lte{Column: "expires_at", Value: now})).
			Update("secret", "")
		scrubbed = res.RowsAffected
		return res.Error
	})
	if err != nil {
		return nil, fmt.Errorf("unable to scrub inactive trigger secrets: %w", err)
	}
	a.mw.Count("general.trigger_secret_cleanup.scrubbed", scrubbed, nil)
	return &ScrubInactiveTriggerSecretsResponse{Scrubbed: scrubbed}, nil
}

package activities

import (
	"context"
	"fmt"
)

type GetFeatureFlagRatiosRequest struct{}

// FeatureFlagRatio is a single (feature, enabled-ratio) row aggregated across
// all non-deleted orgs. ratio is the fraction of orgs with the feature enabled
// (0.0..1.0).
type FeatureFlagRatio struct {
	Feature string  `gorm:"column:feature"`
	Ratio   float64 `gorm:"column:ratio"`
}

// GetFeatureFlagRatios returns the fraction of non-deleted orgs that have each
// feature flag enabled. Used by the general metrics workflow to surface flag
// lifecycle (fully rolled out, fully off, or actively split) to Datadog.
//
// @temporal-gen-v2 activity
func (a *Activities) GetFeatureFlagRatios(ctx context.Context, req GetFeatureFlagRatiosRequest) ([]FeatureFlagRatio, error) {
	var rows []FeatureFlagRatio

	res := a.db.WithContext(ctx).
		Raw(`
			SELECT
				f.key AS feature,
				AVG(CASE WHEN f.value = 'true' THEN 1.0 ELSE 0.0 END) AS ratio
			FROM orgs o, jsonb_each_text(COALESCE(o.features, '{}'::jsonb)) f
			WHERE o.deleted_at = 0
			GROUP BY f.key
		`).
		Scan(&rows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to query feature flag ratios: %w", res.Error)
	}

	return rows, nil
}

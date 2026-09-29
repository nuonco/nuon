package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type featureFlagRow struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Default          bool   `json:"default"`
	Forced           bool   `json:"forced"`
	EffectiveDefault bool   `json:"effective_default"`
	EnabledCount     int    `json:"enabled_count"`
	UnsetCount       int    `json:"unset_count"`
	DriftCount       int    `json:"drift_count"`
}

func (s *service) FeatureFlags(c *gin.Context) {
	ctx := c.Request.Context()

	flags, totalOrgs, err := s.getFeatureFlags(ctx)
	if err != nil {
		s.l.Error("failed to get feature flags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feature flags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"flags":      flags,
		"total_orgs": totalOrgs,
	})
}

func (s *service) getFeatureFlags(ctx context.Context) ([]featureFlagRow, int64, error) {
	var totalOrgs int64
	if err := s.readDB().WithContext(ctx).Model(&app.Org{}).Count(&totalOrgs).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count orgs: %w", err)
	}

	type countRow struct {
		Name       string
		TrueCount  int
		FalseCount int
	}
	var counts []countRow
	if err := s.readDB().WithContext(ctx).Raw(
		`SELECT f.key AS name,
		        COUNT(*) FILTER (WHERE f.value = 'true') AS true_count,
		        COUNT(*) FILTER (WHERE f.value = 'false') AS false_count
		 FROM orgs o, jsonb_each_text(o.features) f
		 WHERE o.deleted_at = 0 AND jsonb_typeof(o.features) = 'object'
		 GROUP BY f.key`,
	).Scan(&counts).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count feature values: %w", err)
	}

	explicit := make(map[string]countRow, len(counts))
	for _, c := range counts {
		explicit[c.Name] = c
	}

	defaults := app.DefaultFeatures()
	forcedFeatures := app.ForcedFeatures()
	features := app.GetFeaturesWithDescriptions()

	rows := make([]featureFlagRow, 0, len(features))
	for i := len(features) - 1; i >= 0; i-- {
		f := features[i]
		def := defaults[app.OrgFeature(f.Name)]
		forced := forcedFeatures[f.Name]
		effective := def || forced

		stored := explicit[f.Name]
		unset := int(totalOrgs) - stored.TrueCount - stored.FalseCount
		if unset < 0 {
			unset = 0
		}

		enabledCount, drift := stored.TrueCount, stored.TrueCount
		if effective {
			enabledCount += unset
			drift = stored.FalseCount
		}
		if forced {
			enabledCount, drift = int(totalOrgs), 0
		}

		rows = append(rows, featureFlagRow{
			Name:             f.Name,
			Description:      f.Description,
			Default:          def,
			Forced:           forced,
			EffectiveDefault: effective,
			EnabledCount:     enabledCount,
			UnsetCount:       unset,
			DriftCount:       drift,
		})
	}

	return rows, totalOrgs, nil
}

func (s *service) featureResolutionJSON() (string, string, string) {
	features := app.GetFeatures()
	forcedFeatures := app.ForcedFeatures()
	defaults := app.DefaultFeatures()

	names := make([]string, 0, len(features))
	forced := make(map[string]string)
	values := make(map[string]string, len(features))
	for _, f := range features {
		name := string(f)
		names = append(names, name)
		values[name] = strconv.FormatBool(defaults[f])
		if forcedFeatures[name] {
			forced[name] = "true"
		}
	}

	namesJSON, _ := json.Marshal(names)
	forcedJSON, _ := json.Marshal(forced)
	valuesJSON, _ := json.Marshal(values)
	return string(namesJSON), string(forcedJSON), string(valuesJSON)
}

func (s *service) effectiveFeatureDefault(name string) bool {
	if app.ForcedFeatures()[name] {
		return true
	}
	return app.DefaultFeatures()[app.OrgFeature(name)]
}

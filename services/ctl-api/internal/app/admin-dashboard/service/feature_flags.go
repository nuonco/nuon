package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type featureFlagRow struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Default is the value a new org gets from the code default, before the
	// deployment's auto_enabled_features override.
	Default          bool `json:"default"`
	AutoEnabled      bool `json:"auto_enabled"`
	EffectiveDefault bool `json:"effective_default"`
	// EnabledCount counts orgs resolved against the default, so orgs predating
	// the flag are counted as whatever the default says.
	EnabledCount int `json:"enabled_count"`
	// UnsetCount is how many orgs have no stored value and fall back to the default.
	UnsetCount int `json:"unset_count"`
	// DriftCount is how many orgs store a value that disagrees with the default.
	DriftCount int `json:"drift_count"`
}

// FeatureFlags returns every active org feature flag with its rollout across orgs.
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
	// jsonb_typeof guards against a scalar `null` features value, which would
	// make jsonb_each_text raise rather than yield zero rows.
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
	autoEnabled := s.autoEnabledFeatures()
	features := app.GetFeaturesWithDescriptions()

	// GetFeatures() is chronological, oldest first, so reverse for newest-first display.
	rows := make([]featureFlagRow, 0, len(features))
	for i := len(features) - 1; i >= 0; i-- {
		f := features[i]
		def := defaults[app.OrgFeature(f.Name)]
		auto := autoEnabled[f.Name]
		effective := def || auto

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

		rows = append(rows, featureFlagRow{
			Name:             f.Name,
			Description:      f.Description,
			Default:          def,
			AutoEnabled:      auto,
			EffectiveDefault: effective,
			EnabledCount:     enabledCount,
			UnsetCount:       unset,
			DriftCount:       drift,
		})
	}

	return rows, totalOrgs, nil
}

// featureDefaultsJSON returns the active flag names as a JSON array and their
// effective defaults as a JSON object of name to "true"/"false", both shaped for
// use as jsonb query parameters.
func (s *service) featureDefaultsJSON() (string, string) {
	features := app.GetFeatures()
	autoEnabled := s.autoEnabledFeatures()
	defaults := app.DefaultFeatures()

	names := make([]string, 0, len(features))
	values := make(map[string]string, len(features))
	for _, f := range features {
		name := string(f)
		names = append(names, name)
		values[name] = strconv.FormatBool(defaults[f] || autoEnabled[name])
	}

	namesJSON, _ := json.Marshal(names)
	valuesJSON, _ := json.Marshal(values)
	return string(namesJSON), string(valuesJSON)
}

// effectiveFeatureDefault is the value an org with no stored entry for the flag
// resolves to: the code default, or forced on by auto_enabled_features.
func (s *service) effectiveFeatureDefault(name string) bool {
	if s.autoEnabledFeatures()[name] {
		return true
	}
	return app.DefaultFeatures()[app.OrgFeature(name)]
}

// autoEnabledFeatures reports the flags this deployment force-enables on new
// orgs regardless of their code default.
func (s *service) autoEnabledFeatures() map[string]bool {
	out := make(map[string]bool)
	if s.cfg == nil {
		return out
	}
	for _, name := range strings.Split(s.cfg.AutoEnabledFeatures, ",") {
		if name = strings.TrimSpace(name); name != "" {
			out[name] = true
		}
	}
	return out
}

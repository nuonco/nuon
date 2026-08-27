package app

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/orgfeatures"

// ForcedFeatures returns the flags this deployment pins on for every org.
func ForcedFeatures() map[string]bool {
	return orgfeatures.Forced()
}

// FeatureForced reports whether the flag is pinned on for every org.
func FeatureForced(feature OrgFeature) bool {
	return orgfeatures.IsForced(string(feature))
}

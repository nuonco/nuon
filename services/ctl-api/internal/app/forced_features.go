package app

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/orgfeatures"

func ForcedFeatures() map[string]bool {
	return orgfeatures.Forced()
}

func FeatureForced(feature OrgFeature) bool {
	return orgfeatures.IsForced(string(feature))
}

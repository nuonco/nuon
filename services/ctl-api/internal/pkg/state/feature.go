package state

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// InstallMetadataStateGenV2Key is the install metadata key that can disable state-gen-v2
// for a specific install even when the org-level feature flag is enabled.
// Set to "false" to opt out: metadata["x-nuon-feature-state-v2"] = "false"
const InstallMetadataStateGenV2Key = "x-nuon-feature-state-v2"

// UseStateGenV2 returns whether v2 state generation should be used for a given install.
// org-level feature flag is the default; install metadata can override it to false.
func UseStateGenV2(orgFeatureEnabled bool, installMetadata pgtype.Hstore) bool {
	if !orgFeatureEnabled {
		return false
	}
	val, ok := installMetadata[InstallMetadataStateGenV2Key]
	if !ok || val == nil {
		return true
	}
	return *val != "false"
}

package plantypes

import (
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

type ContainerImagePullPlan struct {
	Image string `json:"image"`
	Tag   string `json:"tag"`

	// UpdatePolicy is an optional Masterminds-compatible semver constraint
	// (e.g. "~1.25.0") propagated from the user's component config. When
	// non-empty, the runner lists tags from the source registry, filters to
	// those satisfying the constraint, and selects the highest matching
	// tag at build time. Tag is then ignored as the source ref.
	//
	// Empty for components that don't use update_policy.
	UpdatePolicy string `json:"update_policy,omitempty"`

	RepoCfg *configs.OCIRegistryRepository `json:"repo_config"`

	// PreviousSourceDigest is the SourceDigest of the most recent prior Active
	// ComponentBuild for the same component, used by the runner as a dedup
	// hint. When the resolver returns a manifest descriptor whose digest matches
	// this value, the runner skips the Copy step and reports NoOp=true.
	//
	// Empty when there is no prior active build, or when the prior build has
	// no SourceDigest recorded (legacy builds).
	PreviousSourceDigest string `json:"previous_source_digest,omitempty"`
}

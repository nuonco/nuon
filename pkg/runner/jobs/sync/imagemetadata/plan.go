package imagemetadata

import "github.com/nuonco/nuon/pkg/plugins/configs"

// FetchImageMetadataPlan is the local plan type matching the server-side plan.
// We define it locally to avoid import cycle issues with plantypes package.
type FetchImageMetadataPlan struct {
	Registry *configs.OCIRegistryRepository `json:"registry" validate:"required"`
	Tag      string                         `json:"tag" validate:"required"`

	IncludeIndex                bool `json:"include_index"`
	IncludeAttestationManifests bool `json:"include_attestation_manifests"`
	IncludeAttestationLayers    bool `json:"include_attestation_layers"`
}

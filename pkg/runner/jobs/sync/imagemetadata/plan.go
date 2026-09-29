package imagemetadata

import "github.com/nuonco/nuon/pkg/plugins/configs"

type FetchImageMetadataPlan struct {
	Registry *configs.OCIRegistryRepository `json:"registry" validate:"required"`
	Tag      string                         `json:"tag" validate:"required"`

	IncludeIndex                bool `json:"include_index"`
	IncludeAttestationManifests bool `json:"include_attestation_manifests"`
	IncludeAttestationLayers    bool `json:"include_attestation_layers"`
}

package plantypes

import "github.com/nuonco/nuon/pkg/plugins/configs"

// FetchImageMetadataPlan defines the plan for fetching image metadata from an OCI registry.
type FetchImageMetadataPlan struct {
	// Registry configuration for the source image
	Registry *configs.OCIRegistryRepository `json:"registry" validate:"required"`

	// Tag is the image tag to fetch metadata for
	Tag string `json:"tag" validate:"required"`

	// Options for metadata fetching
	IncludeIndex                bool `json:"include_index"`
	IncludeAttestationManifests bool `json:"include_attestation_manifests"`
	IncludeAttestationLayers    bool `json:"include_attestation_layers"`

	MinSandboxMode
}

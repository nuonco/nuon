package plantypes

type TerraformBuildPlan struct {
	Labels map[string]string

	// VendorProviders enables build-time vendoring of terraform providers
	// via `terraform providers mirror`. Gated by the
	// `terraform-provider-mirror` org feature flag in ctl-api so we can
	// roll the change out gradually without coupling install-runner
	// behaviour to the flag (the install runner auto-detects whether a
	// mirror is present in the OCI artifact).
	VendorProviders bool `json:"vendor_providers,omitempty"`

	// TerraformVersion is the version of the terraform CLI the build runner
	// should install in order to vendor providers via
	// `terraform providers mirror`. When empty the build runner falls back
	// to a sane default. Should mirror the version configured for the
	// component's deploy plan so init resolves the same provider bytes the
	// build vendored. Only consulted when VendorProviders is true.
	TerraformVersion string `json:"terraform_version,omitempty"`
}

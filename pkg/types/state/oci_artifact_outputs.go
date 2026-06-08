package state

type OCIArtifactOutputs struct {
	// Repository is the canonical pull reference for the synced artifact.
	// When the runner has resolved a manifest digest, Repository is
	// rewritten to the digest-pinned form (`<full_repo>@sha256:<digest>`)
	// so user templates of the form `image: "{{.outputs.image.repository}}"`
	// are automatically digest-pinned without any user opt-in.
	//
	// Without a resolved digest the original bare-repository form is
	// preserved.
	Repository string `mapstructure:"repository"`
	// Tag is the resolved image tag. Intentionally cleared to `""` whenever
	// Repository has been rewritten to a digest-pinned form, so legacy
	// `image: "{{.repository}}:{{.tag}}"` templates surface as obviously
	// invalid trailing-colon refs (rather than silently deploying a mutable
	// tag). Use DisplayTag for the human-friendly tag.
	Tag string `mapstructure:"tag"`
	// Ref is the canonical digest-pinned pull reference, kept as a stable
	// alias of Repository for clarity:
	//
	//   <login_server>/<repository>@sha256:<digest>
	//
	// Empty when the sync runner did not record a manifest digest.
	Ref string `mapstructure:"ref"`
	// DisplayTag is the human-friendly tag the runner resolved (e.g.
	// "1.25.5") for use in plan output, dashboards, and any user template
	// that wants the tag for display purposes only. Always available even
	// when Tag is intentionally cleared for digest-pinning.
	DisplayTag   string            `mapstructure:"display_tag"`
	MediaType    string            `mapstructure:"media_type"`
	Digest       string            `mapstructure:"digest"`
	Size         int64             `mapstructure:"size"`
	URLs         []string          `mapstructure:"urls"`
	Annotations  map[string]string `mapstructure:"annotations"`
	ArtifactType string            `mapstructure:"artifact_type"`
	Platform     Platform          `mapstructure:"platform"`
}

type Platform struct {
	Architecture string   `mapstructure:"architecture"`
	OS           string   `mapstructure:"os"`
	OSVersion    string   `mapstructure:"os_version"`
	Variant      string   `mapstructure:"variant"`
	OSFeatures   []string `mapstructure:"os_features"`
}

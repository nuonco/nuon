package state

type OCIArtifactOutputs struct {
	// Repository is the bare repository of the synced artifact (no tag, no
	// digest). Long-standing public output contract: user templates compose
	// it as `image: "{{.repository}}:{{.tag}}"`, so it must never carry a
	// digest suffix — use Ref for a digest-pinned pull reference.
	Repository string `mapstructure:"repository"`
	// Tag is the resolved image tag, composed with Repository by user
	// templates. Never cleared: charts default an empty tag to their
	// appVersion, which silently produces invalid refs when combined with
	// any non-bare repository.
	Tag string `mapstructure:"tag"`
	// Ref is the canonical digest-pinned pull reference for templates that
	// opt in to digest-pinned pulls:
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

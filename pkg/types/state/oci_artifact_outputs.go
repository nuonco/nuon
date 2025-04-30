package state

type OCIArtifactOutputs struct {
	Tag          string            `mapstructure:"tag"`
	Repository   string            `mapstructure:"repository"`
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

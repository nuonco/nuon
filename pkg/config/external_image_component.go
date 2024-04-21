package config

type AWSECRConfig struct {
	IAMRoleARN string `mapstructure:"iam_role_arn,omitempty" toml:"iam_role_arn"`
	AWSRegion  string `mapstructure:"region,omitempty" toml:"region"`
	ImageURL   string `mapstructure:"image_url,omitempty" toml:"image_url"`
	Tag        string `mapstructure:"tag,omitempty" toml:"tag"`
}

type PublicImageConfig struct {
	ImageURL string `mapstructure:"image_url,omitempty" toml:"image_url"`
	Tag      string `mapstructure:"tag,omitempty" toml:"tag"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type ExternalImageComponentConfig struct {
	Name         string   `mapstructure:"name,omitempty"`
	Dependencies []string `mapstructure:"dependencies,omitempty"`

	AWSECRImageConfig *AWSECRConfig      `mapstructure:"aws_ecr,omitempty"`
	PublicImageConfig *PublicImageConfig `mapstructure:"public,omitempty"`
}

func (t *ExternalImageComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return resource, nil
}

func (t *ExternalImageComponentConfig) parse() error {
	return nil
}

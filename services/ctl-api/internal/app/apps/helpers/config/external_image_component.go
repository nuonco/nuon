package config

type AWSECRConfig struct {
	IAMRoleARN string `mapstructure:"iam_role_arn" toml:"iam_role_arn"`
	AWSRegion  string `mapstructure:"region" toml:"region"`
	ImageURL   string `mapstructure:"image_url" toml:"image_url"`
	Tag        string `mapstructure:"tag" toml:"tag"`
}

type PublicImageConfig struct {
	ImageURL string `mapstructure:"image_url" toml:"image_url"`
	Tag      string `mapstructure:"tag" toml:"tag"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type ExternalImageComponentConfig struct {
	Name         string   `mapstructure:"name"`
	Dependencies []string `mapstructure:"dependencies"`

	AWSECRImageConfig *AWSECRConfig      `mapstructure:"aws_ecr"`
	PublicImageConfig *PublicImageConfig `mapstructure:"public"`
}

func (t *ExternalImageComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return resource, nil
}

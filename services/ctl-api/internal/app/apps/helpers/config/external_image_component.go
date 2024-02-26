package config

type AWSECRConfig struct {
	IAMRoleARN string `mapstructure:"iam_role_arn" toml:"iam_role_arn"`
	AWSRegion  string `mapstructure:"aws_region" toml:"aws_region"`
}

type ExternalImageComponentConfig struct {
	ImageURL string `mapstructure:"image_url" toml:"image_url"`
	Tag      string `mapstructure:"tag" toml:"tag"`

	AWSECRImageConfig *AWSECRConfig `mapstructure:"aws_ecr_image_config" toml:"aws_ecr_image_config"`
}

func (t *ExternalImageComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return resource, nil
}

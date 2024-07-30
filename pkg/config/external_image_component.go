package config

type AWSECRConfig struct {
	IAMRoleARN string `mapstructure:"iam_role_arn,omitempty" jsonschema:"required"`
	AWSRegion  string `mapstructure:"region,omitempty" jsonschema:"required"`
	ImageURL   string `mapstructure:"image_url,omitempty" jsonschema:"required"`
	Tag        string `mapstructure:"tag,omitempty" jsonschema:"required"`
}

type PublicImageConfig struct {
	ImageURL string `mapstructure:"image_url,omitempty" jsonschema:"required" `
	Tag      string `mapstructure:"tag,omitempty" jsonschema:"required"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type ExternalImageComponentConfig struct {
	MinComponent

	Name         string   `mapstructure:"name,omitempty" jsonschema:"required"`
	Dependencies []string `mapstructure:"dependencies,omitempty"`

	AWSECRImageConfig *AWSECRConfig      `mapstructure:"aws_ecr,omitempty" jsonschema:"oneof_required=public"`
	PublicImageConfig *PublicImageConfig `mapstructure:"public,omitempty" jsonschema:"oneof_required=aws_ecr"`
}

func (t *ExternalImageComponentConfig) Validate() error {
	return nil
}

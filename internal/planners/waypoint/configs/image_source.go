package configs

import "github.com/go-playground/validator/v10"

// PublicImageSource is used as the source for a public image, that does not require authentication.
type PublicImageSource struct {
	Image string `validate:"required"`
	Tag   string `validate:"required"`
}

func (p PublicImageSource) validate(v *validator.Validate) error {
	return v.Struct(v)
}

// PrivateImageSource is used as a private source image for doing a docker-pull. This is used in different contexts,
// such as pulling a prebuilt external image or syncing an image into an install.
//
// Note(jm): essentially, we need to do this https://docs.aws.amazon.com/AmazonECR/latest/userguide/registry_auth.html
type PrivateImageSource struct {
	// this is the output of aws ecr get-login-password --region region | docker login --username AWS
	// --password-stdin aws_account_id.dkr.ecr.region.amazonaws.com
	RegistryToken string `validate:"required"`

	// this should be the registry uri: aws_account_id.dkr.ecr.region.amazonaws.com
	ServerAddress string `validate:"required"`

	// this should be AWS
	Username string `validate:"required"`

	Image string `validate:"required"`
	Tag   string `validate:"required"`
}

func (p PrivateImageSource) validate(v *validator.Validate) error {
	return v.Struct(v)
}

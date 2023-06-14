package sandbox

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
)

type TerraformOutputs struct {
	ClusterID       string `mapstructure:"cluster_name" validate:"required"`
	ClusterEndpoint string `mapstructure:"cluster_endpoint" validate:"required"`
	ClusterCA       string `mapstructure:"cluster_certificate_authority_data" validate:"required"`
	OdrIAMRoleArn   string `mapstructure:"odr_iam_role_arn" json:"odr_iam_role_arn" faker:"len=25" validate:"required"`

	// export other values so they can be written into metadata
	ClusterArn                      string `mapstructure:"cluster_arn"`
	ClusterCertificateAuthorityData string `mapstructure:"cluster_certificate_authority_data"`
	ClusterName                     string `mapstructure:"cluster_name"`
	ClusterPlatformVersion          string `mapstructure:"cluster_platform_version"`
	ClusterStatus                   string `mapstructure:"cluster_status"`
	EcrRegistryID                   string `mapstructure:"ecr_registry_id"`
	EcrRegistryArn                  string `mapstructure:"ecr_registry_arn"`
	EcrRegistryURL                  string `mapstructure:"ecr_registry_url"`
}

func (t *TerraformOutputs) Validate() error {
	validate := validator.New()
	return validate.Struct(t)
}

func ParseTerraformOutputs(outputs *structpb.Struct) (TerraformOutputs, error) {
	m := outputs.AsMap()

	var tfOutputs TerraformOutputs
	if err := mapstructure.Decode(m, &tfOutputs); err != nil {
		return tfOutputs, fmt.Errorf("invalid terraform outputs: %w", err)
	}

	err := tfOutputs.Validate()
	if err != nil {
		return tfOutputs, fmt.Errorf("terraform output error: %w", err)
	}

	return tfOutputs, nil
}

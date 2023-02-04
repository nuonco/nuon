package sandbox

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type TerraformOutputs struct {
	ClusterID       string `mapstructure:"cluster_name"`
	ClusterEndpoint string `mapstructure:"cluster_endpoint"`
	ClusterCA       string `mapstructure:"cluster_certificate_authority_data"`
	OdrIAMRoleArn   string `validate:"required" json:"odr_iam_role_arn" faker:"len=25"`
}

type ParseableTerraformOutputs interface {
	map[string]string | map[string]interface{}
}

func ParseTerraformOutputs[T ParseableTerraformOutputs](inpVals T) (TerraformOutputs, error) {
	var tfOutputs TerraformOutputs
	if err := mapstructure.Decode(inpVals, &tfOutputs); err != nil {
		return tfOutputs, fmt.Errorf("invalid terraform outputs: %w", err)
	}

	return tfOutputs, nil
}

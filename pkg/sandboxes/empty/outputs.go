package empty

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
)

type Outputs struct {
	TestString string                 `mapstructure:"test_string"`
	TestNumber int                    `mapstructure:"test_number"`
	TestMap    map[string]interface{} `mapstructure:"test_map"`

	ClusterName                     string `mapstructure:"cluster_name"`
	ClusterEndpoint                 string `mapstructure:"cluster_endpoint"`
	ClusterCertificateAuthorityData string `mapstructure:"cluster_certificate_authority_data"`
}

func (o *Outputs) Validate() error {
	validate := validator.New()
	return validate.Struct(o)
}

func ParseOutputs(outputs *structpb.Struct) (Outputs, error) {
	m := outputs.AsMap()

	var tfOutputs Outputs
	if err := mapstructure.Decode(m, &tfOutputs); err != nil {
		return tfOutputs, fmt.Errorf("invalid terraform outputs: %w", err)
	}

	err := tfOutputs.Validate()
	if err != nil {
		return tfOutputs, fmt.Errorf("terraform output error: %w", err)
	}

	return tfOutputs, nil
}

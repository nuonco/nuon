package sandboxes

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
)

// CommonOutputs are outputs that need to be set on all sandboxes
type CommonOutputs struct {
	// domain outputs
	PublicDomain   DomainOutputs `mapstructure:"public_domain"`
	InternalDomain DomainOutputs `mapstructure:"internal_domain"`
}

func (t *CommonOutputs) Validate() error {
	validate := validator.New()
	return validate.Struct(t)
}

func ParseCommonOutputs(outputs *structpb.Struct) (CommonOutputs, error) {
	m := outputs.AsMap()

	var tfOutputs CommonOutputs
	if err := mapstructure.Decode(m, &tfOutputs); err != nil {
		return tfOutputs, fmt.Errorf("invalid terraform outputs: %w", err)
	}

	err := tfOutputs.Validate()
	if err != nil {
		return tfOutputs, fmt.Errorf("terraform output error: %w", err)
	}

	return tfOutputs, nil
}

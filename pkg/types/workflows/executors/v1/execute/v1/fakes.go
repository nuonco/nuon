package executev1

import (
	"fmt"
	reflect "reflect"

	structpb "google.golang.org/protobuf/types/known/structpb"
)

func fakeTerraformOutputs(v reflect.Value) (interface{}, error) {
	spb, err := structpb.NewStruct(map[string]interface{}{
		"key": "value",
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get new struct: %w", err)
	}

	return &ExecutePlanResponse_TerraformOutputs{
		TerraformOutputs: spb,
	}, nil
}

package installsv1

import (
	reflect "reflect"

	"google.golang.org/protobuf/types/known/structpb"
)

func fakeTerraformOutputs(reflect.Value) (interface{}, error) {
	data := map[string]interface{}{
		"key": "value",
		"obj": map[string]interface{}{
			"key": "value",
		},
	}

	return structpb.NewStruct(data)
}

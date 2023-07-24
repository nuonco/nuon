package awseks

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid"
	"google.golang.org/protobuf/types/known/structpb"
)

func fakeDomain(reflect.Value) (interface{}, error) {
	return fmt.Sprintf("%s.nuon.run", shortid.New()), nil
}

func fakeStringSliceAsInt(reflect.Value) (interface{}, error) {
	return []interface{}{
		generics.GetFakeObj[string](),
		generics.GetFakeObj[string](),
		generics.GetFakeObj[string](),
	}, nil
}

func fakeOutputs(reflect.Value) (interface{}, error) {
	obj := generics.GetFakeObj[TerraformOutputs]()
	var out map[string]interface{}
	if err := mapstructure.Decode(obj, &out); err != nil {
		return nil, fmt.Errorf("unable to decode outputs: %w", err)
	}

	return structpb.NewStruct(out)
}

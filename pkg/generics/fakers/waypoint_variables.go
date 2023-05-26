package fakers

import (
	"reflect"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func fakeWaypointVariables(v reflect.Value) (interface{}, error) {
	return []*planv1.WaypointVariable{
		{
			Name:  "NUON_TEST_KEY_1",
			Value: "NUON_TEST_VALUE_1",
		},
		{
			Name:  "NUON_TEST_KEY_2",
			Value: "NUON_TEST_VALUE_2",
		},
	}, nil
}

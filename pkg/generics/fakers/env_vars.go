package fakers

import (
	"reflect"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func fakeEnvVars(v reflect.Value) (interface{}, error) {
	return []*planv1.EnvVar{
		{
			Type:  planv1.EnvVarType_ENV_VAR_TYPE_DEPLOY,
			Name:  "NUON_TEST_KEY_1",
			Value: "NUON_TEST_VALUE_1",
		},
		{
			Type:  planv1.EnvVarType_ENV_VAR_TYPE_DEPLOY,
			Name:  "NUON_TEST_KEY_2",
			Value: "NUON_TEST_VALUE_2",
		},
	}, nil
}

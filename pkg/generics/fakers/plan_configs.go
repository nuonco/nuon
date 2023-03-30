package fakers

import (
	"reflect"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func fakePlanConfigs(v reflect.Value) (interface{}, error) {
	return &planv1.Configs{
		Configs: []*planv1.Config{
			{
				Actual: &planv1.Config_EnvVar{
					EnvVar: &planv1.EnvVar{
						Type:  planv1.EnvVarType_ENV_VAR_TYPE_DEPLOY,
						Name:  "NUON_TEST_KEY",
						Value: "NUON_TEST_VALUE",
					},
				},
			},
			{
				Actual: &planv1.Config_EnvVar{
					EnvVar: &planv1.EnvVar{
						Type:  planv1.EnvVarType_ENV_VAR_TYPE_RUNNER_JOB,
						Name:  "NUON_TEST_KEY",
						Value: "NUON_TEST_VALUE",
					},
				},
			},
			{
				Actual: &planv1.Config_HelmValue{
					HelmValue: &planv1.HelmValue{
						Type:  planv1.HelmValueType_HELM_VALUE_TYPE_DEFAULT,
						Name:  "test.value",
						Value: "NUON_TEST_VALUE",
					},
				},
			},
			{
				Actual: &planv1.Config_HelmValue{
					HelmValue: &planv1.HelmValue{
						Type:  planv1.HelmValueType_HELM_VALUE_TYPE_BUILTIN_IMAGE,
						Name:  "test.value",
						Value: "NUON_TEST_VALUE",
					},
				},
			},
			{
				Actual: &planv1.Config_HelmValue{
					HelmValue: &planv1.HelmValue{
						Type:  planv1.HelmValueType_HELM_VALUE_TYPE_BUILTIN_IMAGE_TAG,
						Name:  "test.value",
						Value: "NUON_TEST_VALUE",
					},
				},
			},
		},
	}, nil
}

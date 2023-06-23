package planv1

import (
	"reflect"

	"github.com/go-faker/faker/v4"
)

func fakePlanConfigs(v reflect.Value) (interface{}, error) {
	return []*Config{
		{
			Actual: &Config_EnvVar{
				EnvVar: &EnvVar{
					Type:  EnvVarType_ENV_VAR_TYPE_DEPLOY,
					Name:  "NUON_TEST_KEY",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
		{
			Actual: &Config_EnvVar{
				EnvVar: &EnvVar{
					Type:  EnvVarType_ENV_VAR_TYPE_BUILD,
					Name:  "NUON_TEST_KEY",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
		{
			Actual: &Config_HelmValue{
				HelmValue: &HelmValue{
					Type:  HelmValueType_HELM_VALUE_TYPE_DEFAULT,
					Name:  "test.value",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
		{
			Actual: &Config_HelmValue{
				HelmValue: &HelmValue{
					Type:  HelmValueType_HELM_VALUE_TYPE_BUILTIN_IMAGE,
					Name:  "test.value",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
		{
			Actual: &Config_HelmValue{
				HelmValue: &HelmValue{
					Type:  HelmValueType_HELM_VALUE_TYPE_BUILTIN_IMAGE_TAG,
					Name:  "test.value",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
	}, nil
}

func fakeEnvVars(v reflect.Value) (interface{}, error) {
	return []*EnvVar{
		{
			Type:  EnvVarType_ENV_VAR_TYPE_DEPLOY,
			Name:  "NUON_TEST_KEY_1",
			Value: "NUON_TEST_VALUE_1",
		},
		{
			Type:  EnvVarType_ENV_VAR_TYPE_DEPLOY,
			Name:  "NUON_TEST_KEY_2",
			Value: "NUON_TEST_VALUE_2",
		},
	}, nil
}

func fakeSandboxInputAccountSettings(v reflect.Value) (interface{}, error) {
	return &SandboxInput_Aws{
		Aws: &AWSSettings{
			Region:    "us-west-2",
			AccountId: "543425801867",
			RoleArn:   "arn:aws:iam::543425801867:role/nuon-workers-canary-prod-install-access",
		},
	}, nil
}

func fakeWaypointVariables(v reflect.Value) (interface{}, error) {
	return []*WaypointVariable{
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

func init() {
	_ = faker.AddProvider("planConfigs", fakePlanConfigs)
	_ = faker.AddProvider("envVars", fakeEnvVars)
	_ = faker.AddProvider("sandboxInputAccountSettings", fakeSandboxInputAccountSettings)
	_ = faker.AddProvider("waypointVariables", fakeWaypointVariables)
}

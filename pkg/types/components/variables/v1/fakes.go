package variablesv1

import (
	"reflect"

	"github.com/go-faker/faker/v4"
)

func init() {
	_ = faker.AddProvider("variables", fakeVariables)
	_ = faker.AddProvider("envVars", fakeEnvVars)
	_ = faker.AddProvider("helmValues", fakeHelmValues)
	_ = faker.AddProvider("terraformVariables", fakeTerraformVariables)
	_ = faker.AddProvider("waypointVariables", fakeWaypointVariables)
}

func fakeVariables(v reflect.Value) (interface{}, error) {
	return []*Variable{
		{
			Actual: &Variable_EnvVar{
				EnvVar: &EnvVar{
					Name:  "NUON_TEST_KEY",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
		{
			Actual: &Variable_EnvVar{
				EnvVar: &EnvVar{
					Name:  "NUON_TEST_KEY",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
		{
			Actual: &Variable_HelmValue{
				HelmValue: &HelmValue{
					Name:  "test.value",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
		{
			Actual: &Variable_HelmValue{
				HelmValue: &HelmValue{
					Name:  "test.value",
					Value: "NUON_TEST_VALUE",
				},
			},
		},
		{
			Actual: &Variable_HelmValue{
				HelmValue: &HelmValue{
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
			Name:  "NUON_TEST_KEY_1",
			Value: "NUON_TEST_VALUE_1",
		},
		{
			Name:  "NUON_TEST_KEY_2",
			Value: "NUON_TEST_VALUE_2",
		},
	}, nil
}

func fakeHelmValues(v reflect.Value) (interface{}, error) {
	return []*HelmValue{
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

func fakeTerraformVariables(v reflect.Value) (interface{}, error) {
	return []*TerraformVariable{
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

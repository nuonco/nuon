package variablesv1

import (
	"reflect"

	structpb "google.golang.org/protobuf/types/known/structpb"
)

func fakeIntermediateData(reflect.Value) (interface{}, error) {
	data := map[string]interface{}{
		"key": "value",
		"obj": map[string]interface{}{
			"key": "value",
		},
	}

	return structpb.NewStruct(data)
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

func fakeInstallInputs(v reflect.Value) (interface{}, error) {
	return []*InstallInput{
		{
			Name:  "nuon_test_input_1",
			Value: "NUON_TEST_VALUE_1",
		},
		{
			Name:  "nuon_test_input_2",
			Value: "NUON_TEST_VALUE_2",
		},
	}, nil
}

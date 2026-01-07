package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildInputGroupsFromInputs(t *testing.T) {
	// Create a test logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// Create a mock service with logger
	s := &service{
		l: logger,
	}

	t.Run("basic input group creation", func(t *testing.T) {
		appInputs := []app.AppInput{
			{
				Name:      "input1",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group1",
				},
			},
			{
				Name:      "input2",
				Required:  false,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group1",
				},
			},
		}

		value1 := "value1"
		value2 := "value2"
		installInputValues := map[string]*string{
			"input1": &value1,
			"input2": &value2,
		}

		result := s.buildInputGroupsFromInputs(appInputs, installInputValues, logger)

		assert.Len(t, result, 1, "should have one input group")
		assert.Equal(t, "group1", result[0].Group)
		assert.Len(t, result[0].Inputs, 2, "should have two inputs")
		assert.Equal(t, "value1", result[0].Inputs["input1"])
		assert.Equal(t, "value2", result[0].Inputs["input2"])
	})

	t.Run("multiple input groups sorted alphabetically", func(t *testing.T) {
		appInputs := []app.AppInput{
			{
				Name:      "input3",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "zebra_group",
				},
			},
			{
				Name:      "input2",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "apple_group",
				},
			},
			{
				Name:      "input1",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "middle_group",
				},
			},
		}

		value1 := "value1"
		value2 := "value2"
		value3 := "value3"
		installInputValues := map[string]*string{
			"input1": &value1,
			"input2": &value2,
			"input3": &value3,
		}

		result := s.buildInputGroupsFromInputs(appInputs, installInputValues, logger)

		assert.Len(t, result, 3, "should have three input groups")
		// Check alphabetical ordering
		assert.Equal(t, "apple_group", result[0].Group)
		assert.Equal(t, "middle_group", result[1].Group)
		assert.Equal(t, "zebra_group", result[2].Group)
	})

	t.Run("sensitive inputs are filtered out", func(t *testing.T) {
		appInputs := []app.AppInput{
			{
				Name:      "public_input",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group1",
				},
			},
			{
				Name:      "secret_input",
				Required:  true,
				Sensitive: true,
				AppInputGroup: app.AppInputGroup{
					Name: "group1",
				},
			},
		}

		publicValue := "public_value"
		secretValue := "secret_value"
		installInputValues := map[string]*string{
			"public_input": &publicValue,
			"secret_input": &secretValue,
		}

		result := s.buildInputGroupsFromInputs(appInputs, installInputValues, logger)

		assert.Len(t, result, 1, "should have one input group")
		assert.Len(t, result[0].Inputs, 1, "should have only one input (sensitive filtered)")
		assert.Equal(t, "public_value", result[0].Inputs["public_input"])
		assert.NotContains(t, result[0].Inputs, "secret_input", "sensitive input should be excluded")
	})

	t.Run("required input without value gets empty string", func(t *testing.T) {
		appInputs := []app.AppInput{
			{
				Name:      "required_missing",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group1",
				},
			},
			{
				Name:      "optional_missing",
				Required:  false,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group1",
				},
			},
		}

		installInputValues := map[string]*string{}

		result := s.buildInputGroupsFromInputs(appInputs, installInputValues, logger)

		assert.Len(t, result, 1, "should have one input group")
		assert.Len(t, result[0].Inputs, 1, "should have only required input")
		assert.Equal(t, "", result[0].Inputs["required_missing"], "required missing input should have empty string")
		assert.NotContains(t, result[0].Inputs, "optional_missing", "optional missing input should be excluded")
	})

	t.Run("empty input groups are excluded", func(t *testing.T) {
		appInputs := []app.AppInput{
			{
				Name:      "input1",
				Required:  false,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group_with_no_values",
				},
			},
			{
				Name:      "input2",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group_with_values",
				},
			},
		}

		value2 := "value2"
		installInputValues := map[string]*string{
			"input2": &value2,
		}

		result := s.buildInputGroupsFromInputs(appInputs, installInputValues, logger)

		assert.Len(t, result, 1, "should have only one input group")
		assert.Equal(t, "group_with_values", result[0].Group)
	})

	t.Run("nil value handling", func(t *testing.T) {
		appInputs := []app.AppInput{
			{
				Name:      "input_with_nil",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group1",
				},
			},
		}

		installInputValues := map[string]*string{
			"input_with_nil": nil,
		}

		result := s.buildInputGroupsFromInputs(appInputs, installInputValues, logger)

		assert.Len(t, result, 1, "should have one input group")
		assert.Equal(t, "", result[0].Inputs["input_with_nil"], "nil value should be converted to empty string")
	})

	t.Run("empty input list returns empty result", func(t *testing.T) {
		appInputs := []app.AppInput{}
		installInputValues := map[string]*string{}

		result := s.buildInputGroupsFromInputs(appInputs, installInputValues, logger)

		assert.Len(t, result, 0, "should return empty slice for no inputs")
	})

	t.Run("complex scenario with multiple groups and mixed conditions", func(t *testing.T) {
		appInputs := []app.AppInput{
			// Group A inputs
			{
				Name:      "a_required_with_value",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group_a",
				},
			},
			{
				Name:      "a_optional_with_value",
				Required:  false,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group_a",
				},
			},
			{
				Name:      "a_sensitive",
				Required:  true,
				Sensitive: true,
				AppInputGroup: app.AppInputGroup{
					Name: "group_a",
				},
			},
			// Group B inputs
			{
				Name:      "b_required_missing",
				Required:  true,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group_b",
				},
			},
			{
				Name:      "b_optional_missing",
				Required:  false,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group_b",
				},
			},
			// Group C (all optional, all missing - should be excluded)
			{
				Name:      "c_optional_missing",
				Required:  false,
				Sensitive: false,
				AppInputGroup: app.AppInputGroup{
					Name: "group_c",
				},
			},
		}

		aReqValue := "a_required_value"
		aOptValue := "a_optional_value"
		aSensValue := "a_sensitive_value"
		installInputValues := map[string]*string{
			"a_required_with_value": &aReqValue,
			"a_optional_with_value": &aOptValue,
			"a_sensitive":           &aSensValue,
		}

		result := s.buildInputGroupsFromInputs(appInputs, installInputValues, logger)

		// Should have 2 groups (group_a and group_b), group_c excluded because empty
		assert.Len(t, result, 2, "should have two input groups")

		// Check group_a
		groupA := findGroupByName(result, "group_a")
		require.NotNil(t, groupA, "group_a should exist")
		assert.Len(t, groupA.Inputs, 2, "group_a should have 2 inputs (sensitive excluded)")
		assert.Equal(t, "a_required_value", groupA.Inputs["a_required_with_value"])
		assert.Equal(t, "a_optional_value", groupA.Inputs["a_optional_with_value"])
		assert.NotContains(t, groupA.Inputs, "a_sensitive", "sensitive input should be excluded")

		// Check group_b
		groupB := findGroupByName(result, "group_b")
		require.NotNil(t, groupB, "group_b should exist")
		assert.Len(t, groupB.Inputs, 1, "group_b should have 1 input (required with empty string)")
		assert.Equal(t, "", groupB.Inputs["b_required_missing"], "required missing should have empty string")
		assert.NotContains(t, groupB.Inputs, "b_optional_missing", "optional missing should be excluded")

		// group_c should not exist
		groupC := findGroupByName(result, "group_c")
		assert.Nil(t, groupC, "group_c should not exist (all inputs filtered out)")
	})
}

// Helper function to find a group by name in the result slice
func findGroupByName(groups []config.InputGroup, name string) *config.InputGroup {
	for i := range groups {
		if groups[i].Group == name {
			return &groups[i]
		}
	}
	return nil
}

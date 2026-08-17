package helpers_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

func inputStateCfg() *app.AppConfig {
	return &app.AppConfig{
		InputConfig: app.AppInputConfig{
			AppInputs: []app.AppInput{
				{Name: "cluster_endpoint_public_access", Type: app.AppInputTypeBool, Default: "false"},
				{Name: "cluster_version", Type: app.AppInputTypeString, Default: "1.36"},
				{Name: "root_domain", Type: app.AppInputTypeString},
			},
		},
	}
}

func strPtr(s string) *string { return &s }

// ValuesRedacted carries every declared input, unset ones included as "". Keying
// the default fallback off presence alone resolved those to "" and dropped the
// default, so a declared bool reached terraform as "" and failed plan with
// "a bool is required".
func TestToInputStateAppliesDefaultForEmptyValue(t *testing.T) {
	inputs := &app.InstallInputs{
		ValuesRedacted: pgtype.Hstore{
			"cluster_endpoint_public_access": strPtr(""),
			"cluster_version":                strPtr("1.36"),
			"root_domain":                    strPtr(""),
		},
	}

	is := helpers.ToInputState(inputs, inputStateCfg(), true)
	require.NotNil(t, is)

	assert.Equal(t, "false", is.Inputs["cluster_endpoint_public_access"])
	assert.Equal(t, "1.36", is.Inputs["cluster_version"])
	// No default declared, so an unset input stays empty rather than inventing one.
	assert.Equal(t, "", is.Inputs["root_domain"])
}

func TestToInputStateAppliesDefaultForMissingKey(t *testing.T) {
	inputs := &app.InstallInputs{
		Values: pgtype.Hstore{
			"cluster_version": strPtr("1.40"),
		},
	}

	is := helpers.ToInputState(inputs, inputStateCfg(), false)
	require.NotNil(t, is)

	assert.Equal(t, "false", is.Inputs["cluster_endpoint_public_access"])
	assert.Equal(t, "1.40", is.Inputs["cluster_version"], "an explicit value must win over the default")
}

// An install that sets no inputs at all still needs its declared defaults; the
// previous empty-map short circuit returned nil and dropped every one of them.
func TestToInputStateAppliesDefaultsWithNoValues(t *testing.T) {
	is := helpers.ToInputState(&app.InstallInputs{}, inputStateCfg(), false)
	require.NotNil(t, is)

	assert.Equal(t, "false", is.Inputs["cluster_endpoint_public_access"])
	assert.Equal(t, "1.36", is.Inputs["cluster_version"])
}

func TestToInputStateNilWhenNothingDeclared(t *testing.T) {
	assert.Nil(t, helpers.ToInputState(&app.InstallInputs{}, &app.AppConfig{}, false))
	assert.Nil(t, helpers.ToInputState(nil, inputStateCfg(), false))
}

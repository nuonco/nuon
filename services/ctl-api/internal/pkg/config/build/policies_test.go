package build

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

const testRego = `package nuon

deny contains msg if {
	msg := "nope"
}
`

func TestPoliciesConfigApplyOnTeardown(t *testing.T) {
	no := false

	t.Run("apply_on_teardown is carried from config onto the model", func(t *testing.T) {
		cfg := &config.PoliciesConfig{Policies: []config.AppPolicy{{
			Type:            config.AppPolicyTypeTerraformModule,
			Engine:          config.AppPolicyEngineOPA,
			Name:            "protect-rds",
			Contents:        testRego,
			Components:      []string{"rds"},
			ApplyOnTeardown: &no,
		}}}

		obj, err := PoliciesConfig(PolicyInputsFromConfig(cfg), "app", "cfg")
		require.NoError(t, err)
		require.Len(t, obj.Policies, 1)
		require.Equal(t, &no, obj.Policies[0].ApplyOnTeardown)
	})

	t.Run("apply_on_teardown is rejected for policy types not evaluated on teardown", func(t *testing.T) {
		_, err := PoliciesConfig([]PolicyInput{{
			Type:            config.AppPolicyTypeContainerImage,
			Engine:          config.AppPolicyEngineOPA,
			Name:            "signed",
			Contents:        testRego,
			Components:      []string{"*"},
			ApplyOnTeardown: &no,
		}}, "app", "cfg")
		require.ErrorContains(t, err, "does not support apply_on_teardown")
	})
}

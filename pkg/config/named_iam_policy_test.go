package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedIAMPolicyNameSurvivesJSON(t *testing.T) {
	cfg := PermissionsConfig{
		NamedPolicies: []NamedIAMPolicy{{Name: "{{.nuon.install.id}}-alb-create"}},
	}

	encoded, err := json.Marshal(cfg)
	require.NoError(t, err)

	var decoded PermissionsConfig
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	require.Len(t, decoded.NamedPolicies, 1)
	assert.Equal(t, "{{.nuon.install.id}}-alb-create", decoded.NamedPolicies[0].Name)
}

func TestPermissionsConfigValidateNamedPolicies(t *testing.T) {
	base := func() *PermissionsConfig {
		return &PermissionsConfig{
			ProvisionRole:   &AppAWSIAMRole{Name: "provision", Type: "provision"},
			MaintenanceRole: &AppAWSIAMRole{Name: "maintenance", Type: "maintenance"},
			DeprovisionRole: &AppAWSIAMRole{Name: "deprovision", Type: "deprovision"},
		}
	}

	t.Run("allows omitted named policies", func(t *testing.T) {
		require.NoError(t, base().Validate())
	})

	t.Run("rejects unknown refs", func(t *testing.T) {
		cfg := base()
		cfg.NamedPolicies = []NamedIAMPolicy{{Name: "{{.nuon.install.id}}-alb-create"}}
		cfg.ProvisionRole.NamedPolicies = NamedPolicyRefs([]string{"missing"})
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown named policy")
	})

	t.Run("rejects duplicates", func(t *testing.T) {
		cfg := base()
		cfg.NamedPolicies = []NamedIAMPolicy{
			{Name: "{{.nuon.install.id}}-alb-create"},
			{Name: "{{.nuon.install.id}}-alb-create"},
		}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicated")
	})

	t.Run("accepts a matching ref", func(t *testing.T) {
		cfg := base()
		cfg.NamedPolicies = []NamedIAMPolicy{{Name: "{{.nuon.install.id}}-alb-create"}}
		cfg.ProvisionRole.NamedPolicies = NamedPolicyRefs([]string{"{{.nuon.install.id}}-alb-create"})
		require.NoError(t, cfg.Validate())
	})
}

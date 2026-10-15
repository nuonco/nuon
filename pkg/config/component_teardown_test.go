package config

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestComponentTeardownDecode(t *testing.T) {
	raw := []byte(`
terraform_version = "1.11.3"

[teardown]
retain = [
  "aws_rds_cluster.main",
  "module.buckets",
  "{{ if eq .nuon.inputs.inputs.retain_data \"true\" }}aws_s3_bucket.data{{ end }}",
]
`)

	obj := map[string]any{}
	require.NoError(t, toml.Unmarshal(raw, &obj))

	var out TerraformModuleComponentConfig
	decCfg := DecoderConfig()
	decCfg.Result = &out
	dec, err := mapstructure.NewDecoder(decCfg)
	require.NoError(t, err)
	require.NoError(t, dec.Decode(obj))

	require.NotNil(t, out.Teardown)
	require.Equal(t, []string{
		"aws_rds_cluster.main",
		"module.buckets",
		`{{ if eq .nuon.inputs.inputs.retain_data "true" }}aws_s3_bucket.data{{ end }}`,
	}, out.Teardown.Retain)
	require.NoError(t, out.Validate())
}

func TestComponentTeardownConfigValidate(t *testing.T) {
	tests := map[string]struct {
		teardown *ComponentTeardownConfig
		wantErr  string
	}{
		"nil": {},
		"empty": {
			teardown: &ComponentTeardownConfig{},
		},
		"valid addresses": {
			teardown: &ComponentTeardownConfig{Retain: []string{
				"aws_rds_cluster.main",
				`aws_s3_bucket.data["logs"]`,
				"module.buckets",
			}},
		},
		"templated address is checked once rendered": {
			teardown: &ComponentTeardownConfig{Retain: []string{
				"{{ .nuon.inputs.inputs.retained_resource }}",
			}},
		},
		"invalid template": {
			teardown: &ComponentTeardownConfig{Retain: []string{"{{ if .nuon.inputs.inputs.retain_data }}aws_rds_cluster.main"}},
			wantErr:  "teardown.retain[0]: invalid template",
		},
		"invalid address": {
			teardown: &ComponentTeardownConfig{Retain: []string{"aws_rds_cluster.main", "aws_rds_cluster"}},
			wantErr:  "teardown.retain[1]",
		},
		"data source": {
			teardown: &ComponentTeardownConfig{Retain: []string{"data.aws_ami.ubuntu"}},
			wantErr:  "data source",
		},
		"duplicate address": {
			teardown: &ComponentTeardownConfig{Retain: []string{"module.buckets", "module.buckets"}},
			wantErr:  `teardown.retain[1]: duplicate address "module.buckets"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.teardown.Validate()
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTerraformModuleComponentConfigValidateTeardown(t *testing.T) {
	cfg := TerraformModuleComponentConfig{
		TerraformVersion: "1.11.3",
		Teardown:         &ComponentTeardownConfig{Retain: []string{"aws_rds_cluster"}},
	}

	err := cfg.Validate()
	var cfgErr ErrConfig
	require.ErrorAs(t, err, &cfgErr)
	require.Contains(t, cfgErr.Description, "teardown.retain[0]")
}

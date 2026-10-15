package build

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func TestComponentConnectionTeardown(t *testing.T) {
	tests := map[string]struct {
		comp     *config.Component
		want     *app.ComponentTeardown
		wantCode string
	}{
		"terraform component without teardown": {
			comp: &config.Component{
				Name:            "storage",
				Type:            config.TerraformModuleComponentType,
				TerraformModule: &config.TerraformModuleComponentConfig{TerraformVersion: "1.11.3"},
			},
		},
		"terraform component with empty retain list": {
			comp: &config.Component{
				Name: "storage",
				Type: config.TerraformModuleComponentType,
				TerraformModule: &config.TerraformModuleComponentConfig{
					TerraformVersion: "1.11.3",
					Teardown:         &config.ComponentTeardownConfig{},
				},
			},
		},
		"terraform component with retained resources": {
			comp: &config.Component{
				Name: "storage",
				Type: config.TerraformModuleComponentType,
				TerraformModule: &config.TerraformModuleComponentConfig{
					TerraformVersion: "1.11.3",
					Teardown: &config.ComponentTeardownConfig{Retain: []string{
						"aws_rds_cluster.main",
						"{{ .nuon.inputs.inputs.retained_bucket }}",
					}},
				},
			},
			want: &app.ComponentTeardown{Retain: []string{
				"aws_rds_cluster.main",
				"{{ .nuon.inputs.inputs.retained_bucket }}",
			}},
		},
		"invalid retain address": {
			comp: &config.Component{
				Name: "storage",
				Type: config.TerraformModuleComponentType,
				TerraformModule: &config.TerraformModuleComponentConfig{
					TerraformVersion: "1.11.3",
					Teardown:         &config.ComponentTeardownConfig{Retain: []string{"data.aws_ami.ubuntu"}},
				},
			},
			wantCode: "invalid_teardown",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			in, err := ComponentConnectionInputFromConfig(tc.comp, "cmp-storage", "apc-1", nil)
			require.NoError(t, err)

			ccc, err := ComponentConnection(in)
			if tc.wantCode != "" {
				var userErr stderr.ErrUser
				require.ErrorAs(t, err, &userErr)
				require.Equal(t, tc.wantCode, userErr.Code)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, ccc.Teardown)
		})
	}
}

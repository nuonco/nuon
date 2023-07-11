package buildv1

import (
	"fmt"
	"testing"

	vcsv1 "github.com/powertoolsdev/mono/pkg/types/components/vcs/v1"
	"github.com/stretchr/testify/assert"
)

func TestTerraformModuleConfig_Validate(t *testing.T) {
	tests := map[string]struct {
		cfgFn       func() *TerraformModuleConfig
		errExpected error
	}{
		"happy path": {
			cfgFn: func() *TerraformModuleConfig {
				return &TerraformModuleConfig{
					VcsCfg: &vcsv1.Config{
						Cfg: &vcsv1.Config_PublicGitConfig{
							PublicGitConfig: &vcsv1.PublicGitConfig{
								Repo:      "repo",
								Directory: "dir",
								GitRef:    "main",
							},
						},
					},
				}
			},
		},
		"error no vcs config": {
			cfgFn: func() *TerraformModuleConfig {
				return &TerraformModuleConfig{}
			},
			errExpected: fmt.Errorf("VcsCfg: value is required"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := test.cfgFn()
			err := cfg.Validate()
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

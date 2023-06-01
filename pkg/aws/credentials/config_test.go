package credentials

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := map[string]struct {
		configFn    func() *Config
		errExpected error
	}{
		"happy path - static": {
			configFn: func() *Config {
				cfg := generics.GetFakeObj[*Config]()
				cfg.AssumeRoleConfig = AssumeRoleConfig{}
				return cfg
			},
		},
		"happy path - assume role": {
			configFn: func() *Config {
				cfg := generics.GetFakeObj[*Config]()
				cfg.StaticCredentials = StaticCredentials{}
				return cfg
			},
		},
		"both invalid": {
			configFn: func() *Config {
				return &Config{}
			},
			errExpected: fmt.Errorf("unable to validate"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			cfg := test.configFn()

			err := cfg.Validate(v)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

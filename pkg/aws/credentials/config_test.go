package credentials

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestConfig_parsing(t *testing.T) {
	filename := "test.hcl"

	tests := map[string]struct {
		hcl         string
		output      Config
		errExpected error
	}{
		"happy path - default": {
			hcl: `use_default = true`,
			output: Config{
				UseDefault: true,
			},
			errExpected: nil,
		},
		"happy path - static": {
			hcl: `
static {
	access_key_id = "access-key"
	secret_access_key = "secret-key"
	session_token = "session-token"
}`,

			output: Config{
				Static: &StaticCredentials{
					AccessKeyID:     "access-key",
					SecretAccessKey: "secret-key",
					SessionToken:    "session-token",
				},
			},
			errExpected: nil,
		},
		"happy path - assume": {
			hcl: `
assume_role {
	role_arn = "role-arn"
	session_name = "session-name"
	session_duration_seconds = 60
}`,

			output: Config{
				AssumeRole: &AssumeRoleConfig{
					RoleARN:                "role-arn",
					SessionName:            "session-name",
					SessionDurationSeconds: 60,
				},
			},
			errExpected: nil,
		},
		"happy path - assume default duration": {
			hcl: `
assume_role {
	role_arn = "role-arn"
	session_name = "session-name"
}`,

			output: Config{
				AssumeRole: &AssumeRoleConfig{
					RoleARN:                "role-arn",
					SessionName:            "session-name",
					SessionDurationSeconds: 0,
				},
			},
			errExpected: nil,
		},
		"missing - arn on assume_role": {
			hcl: `
assume_role {
	session_name = "session-name"
	session_duration_seconds = 60
}`,

			errExpected: fmt.Errorf("arn"),
		},
		"missing - session name on assume_role": {
			hcl: `
assume_role {
	role_arn = "role-arn"
	session_duration_seconds = 60
}`,

			errExpected: fmt.Errorf("session_name"),
		},
		"missing - access_key on static": {
			hcl: `
static {
	secret_access_key = "secret-key"
	session_token = "session-token"
}`,

			errExpected: fmt.Errorf("access_key"),
		},
		"missing - secret access_key on static": {
			hcl: `
static {
	access_key_id = "access-key"
	session_token = "session-token"
}`,

			errExpected: fmt.Errorf("access_key"),
		},
		"missing - session token on static": {
			hcl: `
static {
	access_key_id = "access-key"
	secret_access_key = "secret-key"
}`,

			errExpected: fmt.Errorf("session_token"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var cfg Config

			hclFile, diag := hclparse.NewParser().ParseHCL([]byte(test.hcl), filename)
			assert.False(t, diag.HasErrors())

			diag = gohcl.DecodeBody(hclFile.Body, nil, &cfg)
			if test.errExpected != nil {
				assert.True(t, diag.HasErrors())

				for _, err := range diag.Errs() {
					if strings.Contains(fmt.Sprintf("%s", err), test.errExpected.Error()) {
						return
					}
				}

				assert.Fail(t, "error was not contained in any diag errors")
				return
			}

			assert.Equal(t, test.output, cfg)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := map[string]struct {
		configFn    func() *Config
		errExpected error
	}{
		"happy path - default": {
			configFn: func() *Config {
				cfg := generics.GetFakeObj[*Config]()
				cfg.UseDefault = true
				return cfg
			},
		},
		"happy path - static": {
			configFn: func() *Config {
				cfg := generics.GetFakeObj[*Config]()
				cfg.AssumeRole = nil
				return cfg
			},
		},
		"happy path - assume role": {
			configFn: func() *Config {
				cfg := generics.GetFakeObj[*Config]()
				cfg.Static = nil
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

func TestConfig_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		configFn func() *Config
		assertFn func(*testing.T, map[string]interface{})
	}{
		"happy path - use default": {
			configFn: func() *Config {
				return &Config{
					UseDefault: true,
				}
			},
			assertFn: func(t *testing.T, val map[string]interface{}) {
				assert.Equal(t, true, val["use_default"])
			},
		},
		"happy path - static": {
			configFn: func() *Config {
				return &Config{
					Static: &StaticCredentials{
						AccessKeyID:     "access-key",
						SecretAccessKey: "secret-access-key",
						SessionToken:    "session-token",
					},
				}
			},
			assertFn: func(t *testing.T, val map[string]interface{}) {
				static := val["static"].(map[string]interface{})

				assert.Equal(t, 3, len(static))

				assert.Equal(t, "access-key", static["access_key"])
				assert.Equal(t, "secret-access-key", static["secret_key"])
				assert.Equal(t, "session-token", static["token"])
			},
		},
		"happy path - assume_role": {
			configFn: func() *Config {
				return &Config{
					AssumeRole: &AssumeRoleConfig{
						RoleARN:                "role-arn",
						SessionName:            "session-name",
						SessionDurationSeconds: 5,
					},
				}
			},
			assertFn: func(t *testing.T, val map[string]interface{}) {
				assume := val["assume_role"].(map[string]interface{})
				assert.Equal(t, 4, len(assume))

				assert.Equal(t, "role-arn", assume["role_arn"])
				assert.Equal(t, "session-name", assume["session_name"])
				assert.Equal(t, 5.0, assume["session_duration_seconds"])
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := test.configFn()
			byts, err := json.Marshal(cfg)
			assert.NoError(t, err)

			var output map[string]interface{}
			err = json.Unmarshal(byts, &output)
			assert.NoError(t, err)

			test.assertFn(t, output)
		})
	}
}

package credentials

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"

	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

// AssumeRoleConfig is used for assuming an IAM role
type AssumeRoleConfig struct {
	RoleARN                string `cty:"arn" hcl:"role_arn" validate:"required" mapstructure:"role_arn,omitempty" json:"role_arn"`
	SessionName            string `cty:"session_name" hcl:"session_name" validate:"required" mapstructure:"session_name,omitempty" json:"session_name"`
	SessionDurationSeconds int    `cty:"session_duration_seconds" hcl:"session_duration_seconds" mapstructure:"session_duration_seconds,omitempty" json:"session_duration_seconds"`

	// configuration for two stepping before assuming this role
	TwoStepConfig *assumerole.TwoStepConfig `cty:"two_step_config" hcl:"two_step_config" mapstructure:"two_step_config,omitempty" json:"two_step_config"`
}

// StaticCredentials are used to create credentials ahead of time, and pass them around for use. Specifically, we do
// this for creating credentials with an IAM role in our infra, so a plugin can push data back.
type StaticCredentials struct {
	AccessKeyID     string `cty:"access_key_id" hcl:"access_key_id" validate:"required" mapstructure:"access_key,omitempty" json:"access_key_id"`
	SecretAccessKey string `cty:"secret_access_key" hcl:"secret_access_key" validate:"required" mapstructure:"secret_key,omitempty" json:"secret_access_key"`
	SessionToken    string `cty:"session_token" hcl:"session_token" validate:"required" mapstructure:"token,omitempty" json:"session_token"`
}

type Config struct {
	Static     *StaticCredentials `cty:"static,block" hcl:"static,block" mapstructure:"static,omitempty" json:"static"`
	AssumeRole *AssumeRoleConfig  `cty:"assume_role,block" hcl:"assume_role,block" mapstructure:"assume_role,omitempty" json:"assume_role"`
	UseDefault bool               `cty:"use_default,optional" hcl:"use_default,optional" mapstructure:"use_default,omitempty" json:"use_default"`

	// when cache ID is set, these credentials will be reused, up to the duration of the sessionTimeout (or default)
	CacheID string `cty:"cache_id,optional" hcl:"cache_id,optional" json:"cache_id,omitempty" mapstructure:"cache_id,omitempty"`
	Region  string `cty:"region,optional" hcl:"region,optional" mapstructure:"region,omitempty" json:"region"`
}

func (c Config) MarshalJSON() ([]byte, error) {
	var output map[string]interface{}
	if err := mapstructure.Decode(c, &output); err != nil {
		return nil, fmt.Errorf("unable to decode to stringmap: %w", err)
	}

	return json.Marshal(output)
}

func (c *Config) Validate(v *validator.Validate) error {
	if c.UseDefault {
		return nil
	}

	credsErr := v.Struct(c.Static)
	roleErr := v.Struct(c.AssumeRole)
	if credsErr != nil && roleErr != nil {
		return errors.Join(fmt.Errorf("unable to validate credentials: %w", credsErr),
			fmt.Errorf("unable to validate role: %w", roleErr))
	}

	return nil
}

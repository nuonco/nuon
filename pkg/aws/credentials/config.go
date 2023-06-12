package credentials

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

// AssumeRoleConfig is used for assuming an IAM role
type AssumeRoleConfig struct {
	RoleARN                string `hcl:"arn" validate:"required" mapstructure:"role_arn,omitempty"`
	SessionName            string `hcl:"session_name" validate:"required" mapstructure:"session_name,omitempty"`
	SessionDurationSeconds int    `hcl:"session_duration_seconds" mapstructure:"session_duration_seconds,omitempty"`
}

// StaticCredentials are used to create credentials ahead of time, and pass them around for use. Specifically, we do
// this for creating credentials with an IAM role in our infra, so a plugin can push data back.
type StaticCredentials struct {
	AccessKeyID     string `hcl:"access_key_id" validate:"required" mapstructure:"access_key,omitempty"`
	SecretAccessKey string `hcl:"secret_access_key" validate:"required" mapstructure:"secret_key,omitempty"`
	SessionToken    string `hcl:"session_token" validate:"required" mapstructure:"token,omitempty"`
}

type Config struct {
	Static     StaticCredentials `hcl:"static,block" mapstructure:",squash"`
	AssumeRole AssumeRoleConfig  `hcl:"assume_role,block" mapstructure:",squash"`
	UseDefault bool              `hcl:"use_default,optional" mapstructure:"use_default,omitempty"`

	// when cache ID is set, these credentials will be reused, up to the duration of the sessionTimeout (or default)
	CacheID string `hcl:"cache_id,optional" json:"cache_id,omitempty" mapstructure:"cache_id,omitempty"`
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

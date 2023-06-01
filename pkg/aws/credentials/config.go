package credentials

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// AssumeRoleConfig is used for assuming an IAM role
type AssumeRoleConfig struct {
	ARN            string        `hcl:"role_arn,optional" validate:"required"`
	SessionName    string        `hcl:"role_session_name,optional" validate:"required"`
	SessionTimeout time.Duration `hcl:"role_session_name" validate:"optional"`
}

// StaticCredentials are used to create credentials ahead of time, and pass them around for use. Specifically, we do
// this for creating credentials with an IAM role in our infra, so a plugin can push data back.
type StaticCredentials struct {
	AccessKeyID     string `hcl:"access_key_id,optional" validate:"required"`
	SecretAccessKey string `hcl:"secret_access_key,optional" validate:"required"`
	SessionToken    string `hcl:"session_token,optional" validate:"required"`
}

type Config struct {
	StaticCredentials
	AssumeRoleConfig

	// when cache ID is set, these credentials will be reused, up to the duration of the sessionTimeout (or default)
	CacheID string `hcl:"cache_id"`
}

func (c *Config) Validate(v *validator.Validate) error {
	credsErr := v.Struct(c.StaticCredentials)
	roleErr := v.Struct(c.AssumeRoleConfig)

	if credsErr != nil && roleErr != nil {
		return errors.Join(fmt.Errorf("unable to validate credentials: %w", credsErr),
			fmt.Errorf("unable to validate role: %w", roleErr))
	}

	return nil
}

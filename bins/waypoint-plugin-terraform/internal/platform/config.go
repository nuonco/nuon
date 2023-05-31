package platform

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

//
// NOTE(jm): this plugin is not 100% what we want, as the original idea was that we would simply use the `docker-pull`
// plugin to automatically pull a plugin from the vendor's account and push into the local customer's ECR. However, I
// don't think that will work, and all around we probably should have an `oci-pull` plugin do this and avoid the
// complexity that using `docker-pull` brings. However, to get this working we don't need either -- we just pull the
// plugin here, directly.
//

type AWSAuthRoleARN struct {
	RoleARN         string `hcl:"role_arn,optional" validate:"required"`
	RoleSessionName string `hcl:"role_session_name,optional" validate:"required"`
}

type AWSAuthCredentials struct {
	// pre-generated credentials
	AccessKeyID     string `hcl:"access_key_id,optional" validate:"required"`
	SecretAccessKey string `hcl:"secret_access_key,optional" validate:"required"`
	SessionToken    string `hcl:"session_token,optional" validate:"required"`
}

// AWS auth can use a role arn, or creds. Both are passed into the terraform package
type AWSAuth struct {
	AWSAuthRoleARN
	AWSAuthCredentials
}

func (a *AWSAuth) Validate(v *validator.Validate) error {
	credsErr := v.Struct(a.AWSAuthCredentials)
	roleErr := v.Struct(a.AWSAuthRoleARN)

	if credsErr != nil && roleErr != nil {
		return errors.Join(fmt.Errorf("unable to validate creds auth: %w", credsErr),
			fmt.Errorf("unable to validate role auth: %w", roleErr))
	}

	return nil
}

type Config struct {
	Archive struct {
		AuthToken  string `hcl:"auth_token"`
		RegistryID string `hcl:"registry_id"`
		ImageURL   string `hcl:"image_url"`
	} `hcl:"archive,block"`

	TerraformVersion string `hcl:"terraform_version"`

	// auth for the run itself
	RunAuth AWSAuth `hcl:"run_auth,block"`

	// outputs are used to set the outputs after the terraform run
	Backend struct {
		Bucket   string `hcl:"bucket"`
		StateKey string `hcl:"state_key"`
		Prefix   string `hcl:"prefix"`

		Auth AWSAuth `hcl:"aws_auth"`
	} `hcl:"backend,block"`
}

// Config returns a pointer to the config, so that the plugin SDK can serialize into it.
func (p *Platform) Config() (interface{}, error) {
	return &p.Cfg, nil
}

// ConfigSet is a callback for when a configuration is written
func (p *Platform) ConfigSet(config interface{}) error {
	c, ok := config.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type: %T", config)
	}

	if err := c.RunAuth.Validate(p.v); err != nil {
		return fmt.Errorf("unable to validate run auth: %w", err)
	}

	if err := c.Backend.Auth.Validate(p.v); err != nil {
		return fmt.Errorf("unable to validate backend auth: %w", err)
	}

	wkspace, err := p.GetWorkspace()
	if err != nil {
		return fmt.Errorf("unable to create workspace from config: %w", err)
	}
	p.Workspace = wkspace

	if err := p.v.Struct(p); err != nil {
		return fmt.Errorf("unable to validate plugin: %w", err)
	}

	return nil
}

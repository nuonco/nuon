package platform

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

type Config struct {
	Archive struct {
		// NOTE(jm): we can not pull the archive information in from the registry plugin, as waypoint doesn't
		// support that.
		//
		// FWIW, we could just share the code here + config between this and the registry, but that probably
		// needs a bit more refactoring as the build + deploy sides are fairly, fairly decoupled.
		Username    string `hcl:"username"`
		AuthToken   string `hcl:"auth_token"`
		RegistryURL string `hcl:"registry_url"`
		Repo        string `hcl:"repo"`
		Tag         string `hcl:"tag"`
	} `hcl:"archive,block"`

	TerraformVersion string `hcl:"terraform_version"`

	// auth for the run itself
	RunAuth  credentials.Config `hcl:"run_auth,block"`
	PlanOnly bool               `hcl:"plan_only"`

	// outputs are used to set the outputs after the terraform run
	Backend struct {
		Bucket   string             `hcl:"bucket"`
		StateKey string             `hcl:"state_key"`
		Region   string             `hcl:"region"`
		Auth     credentials.Config `hcl:"aws_auth"`
	} `hcl:"backend,block"`

	// Outputs are used to control where the run outputs are synchronized to
	Outputs struct {
		Bucket string             `hcl:"bucket"`
		Auth   credentials.Config `hcl:"aws_auth"`
		Prefix string             `hcl:"prefix"`
	} `hcl:"outputs,block"`

	Labels    map[string]string `hcl:"labels,optional"`
	Variables map[string]string `hcl:"variables,optional"`
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

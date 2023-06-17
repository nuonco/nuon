package platform

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

// Config returns a pointer to the config, so that the plugin SDK can serialize into it.
func (p *Platform) Config() (interface{}, error) {
	return &p.Cfg, nil
}

// ConfigSet is a callback for when a configuration is written
func (p *Platform) ConfigSet(config interface{}) error {
	c, ok := config.(*configs.TerraformDeploy)
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

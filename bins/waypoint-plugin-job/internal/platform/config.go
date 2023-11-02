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
	_, ok := config.(*configs.JobDeploy)
	if !ok {
		return fmt.Errorf("invalid config type: %T", config)
	}

	if err := p.v.Struct(p); err != nil {
		return fmt.Errorf("unable to validate plugin: %w", err)
	}

	return nil
}

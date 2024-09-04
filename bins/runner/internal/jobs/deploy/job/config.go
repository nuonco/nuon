package job

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

// ConfigSet is a callback for when a configuration is written
func (p *handler) ConfigSet(config interface{}) error {
	_, ok := config.(*configs.JobDeploy)
	if !ok {
		return fmt.Errorf("invalid config type: %T", config)
	}

	if err := p.v.Struct(p); err != nil {
		return fmt.Errorf("unable to validate plugin: %w", err)
	}

	return nil
}

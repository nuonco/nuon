package builder

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.config, nil
}

// Implement ConfigurableNotify
func (b *Builder) ConfigSet(config interface{}) error {
	_, ok := config.(*configs.NoopBuild)
	if !ok {
		return fmt.Errorf("expected type NoopBuild")
	}

	return nil
}

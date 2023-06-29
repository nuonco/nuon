package builder

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.cfg, nil
}

// Implement ConfigurableNotify
func (b *Builder) ConfigSet(config interface{}) error {
	_, ok := config.(*configs.OCIArchiveBuild)
	if !ok {
		return fmt.Errorf("expected type BuildConfig")
	}

	return nil
}

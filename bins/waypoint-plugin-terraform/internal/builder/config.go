package builder

import (
	"fmt"
)

type BuildConfig struct {
	OutputName string `hcl:"output_name,optional"`

	Labels    map[string]string `hacl:"labels,optional"`
	Variables map[string]string `hcl:"variables,optional"`
}

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.config, nil
}

// Implement ConfigurableNotify
func (b *Builder) ConfigSet(config interface{}) error {
	_, ok := config.(*BuildConfig)
	if !ok {
		return fmt.Errorf("expected type BuildConfig")
	}

	return nil
}

package builder

import (
	"fmt"
	"os"
)

type BuildConfig struct {
	OutputName string `hcl:"output_name,optional"`
	Source     string `hcl:"source,optional"`

	Labels map[string]string `hacl:"labels,optional"`
}

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.config, nil
}

// Implement ConfigurableNotify
func (b *Builder) ConfigSet(config interface{}) error {
	c, ok := config.(*BuildConfig)
	if !ok {
		return fmt.Errorf("expected type BuildConfig")
	}

	_, err := os.Stat(c.Source)
	if err != nil {
		return fmt.Errorf("source folder does not exist")
	}

	return nil
}

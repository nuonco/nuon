package cmd

import (
	"github.com/powertoolsdev/go-common/config"
	domain "github.com/powertoolsdev/template-go-workers/internal"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")
}

type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	Cfg domain.Config `config:"org"`
}

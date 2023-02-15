package internal

import (
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("env", "local")
}

type Config struct {
	config.Base `config:",squash"`
}

package cmd

import (
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
}

type Config struct {
	config.Base `config:",squash"`
}

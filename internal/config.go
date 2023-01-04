package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_port", "8080")
	config.RegisterDefault("http_address", "0.0.0.0")
}

type Config struct {
	config.Base `config:",squash"`

	// configs for starting and introspecting service
	GitRef      string `config:"git_ref" validate:"required"`
	HTTPPort    string `config:"http_port" validate:"required"`
	HTTPAddress string `config:"http_address" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

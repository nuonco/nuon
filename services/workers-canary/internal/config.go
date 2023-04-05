package deployment

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "canary")
}

type Config struct {
	worker.Config `config:",squash"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

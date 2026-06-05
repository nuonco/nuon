package runbooks

import (
	"github.com/nuonco/nuon/bins/cli/internal/config"
	nuon "github.com/nuonco/nuon/sdks/nuon-go"
)

type Service struct {
	api nuon.Client
	cfg *config.Config
}

func New(api nuon.Client, cfg *config.Config) *Service {
	return &Service{
		api: api,
		cfg: cfg,
	}
}

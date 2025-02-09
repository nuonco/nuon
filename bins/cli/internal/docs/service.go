package docs

import "github.com/powertoolsdev/mono/bins/cli/internal/config"

type Service struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

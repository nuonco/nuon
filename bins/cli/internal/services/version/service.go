package version

import (
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

type Service struct {
	api nuon.Client
	cfg *config.Config
}

func New(api nuon.Client, cfg *config.Config) *Service {
	return &Service{api: api, cfg: cfg}
}

// apiConfigured reports whether the user set api_url. The built-in default is
// not a configuration, so version does not call it.
func (s *Service) apiConfigured() bool {
	if s == nil || s.api == nil || s.cfg == nil || s.cfg.Viper == nil {
		return false
	}
	return strings.TrimSpace(s.cfg.GetString("api_url")) != ""
}

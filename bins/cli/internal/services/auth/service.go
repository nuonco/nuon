package auth

import (
	"net"

	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/bins/cli/internal/config"
)

type Service struct {
	api nuon.Client
	cfg *config.Config

	authBaseURLOverride string
	listenLoopbackFn    func() (net.Listener, error)
}

func New(apiClient nuon.Client, cfg *config.Config) *Service {
	return &Service{
		api: apiClient,
		cfg: cfg,
	}
}

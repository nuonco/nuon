package k8s

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint/pkg/serverclient"
	"google.golang.org/grpc"
)

const (
	defaultHCLogName      string        = "nuon"
	defaultConnectTimeout time.Duration = time.Minute
)

// getClient returns a waypoint client
func (p *k8sClient) getClient(ctx context.Context, addr, token string) (*grpc.ClientConn, error) {
	cfg, err := serverclient.ContextConfig()
	if err != nil {
		return nil, err
	}
	cfg.Server.RequireAuth = false
	cfg.Server.Address = addr
	cfg.Server.Tls = true
	cfg.Server.TlsSkipVerify = true
	cfg.Server.RequireAuth = token != ""
	cfg.Server.AuthToken = token
	primaryOpts := serverclient.FromContextConfig(cfg)

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  defaultHCLogName,
		Level: hclog.LevelFromString("DEBUG"),
	})
	logOpt := serverclient.Logger(appLogger)
	return serverclient.Connect(ctx, primaryOpts, logOpt, serverclient.Timeout(defaultConnectTimeout))
}

package acr

import (
	"context"
	"fmt"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
	"github.com/powertoolsdev/mono/pkg/azure/acr"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

func FetchAccessInfo(ctx context.Context, cfg *configs.OCIRegistryRepository) (*registry.AccessInfo, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, err
	}

	username := ""
	password := ""
	if cfg.OCIAuth != nil {
		l.Info("plan includes oci auth credentials")
		l.Info("using provided oci registry user and token")
		username = cfg.OCIAuth.Username
		password = cfg.OCIAuth.Password
	} else {
		l.Info("plan does not include oci auth credentials")
		l.Info("getting token using azure rbac...")
		token, err := acr.GetRepositoryToken(ctx, cfg.ACRAuth, cfg.LoginServer, l)
		if err != nil {
			return nil, fmt.Errorf("unable to get acr token: %w", err)
		}
		l.Info("got token using auzre rbac")
		username = acr.DefaultACRUsername
		password = token
	}

	return &registry.AccessInfo{
		Image: cfg.Repository,
		Auth: &registry.AccessInfoAuth{
			Username:      username,
			Password:      password,
			ServerAddress: cfg.LoginServer,
		},
	}, nil
}

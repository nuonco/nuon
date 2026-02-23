package gar

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

const (
	garUsername = "oauth2accesstoken"
)

func FetchAccessInfo(ctx context.Context, cfg *configs.OCIRegistryRepository) (*registry.AccessInfo, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, err
	}

	l.Info("getting GAR access token using application default credentials")
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("unable to find default credentials: %w", err)
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to get access token: %w", err)
	}

	l.Info("got GAR access token")
	return &registry.AccessInfo{
		Image: cfg.Repository,
		Auth: &registry.AccessInfoAuth{
			Username:      garUsername,
			Password:      token.AccessToken,
			ServerAddress: cfg.LoginServer,
		},
	}, nil
}

package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	registryAuthUsername = "oauth2accesstoken"
	// must exceed the runner's keep window (20m) so a runner swapping tokens always gets one it will keep
	registryAuthMinRemaining = 30 * time.Minute
	registryAuthMintTimeout  = 10 * time.Second
)

type tokenSourceFactory func(ctx context.Context, serviceAccount string) (oauth2.TokenSource, error)

type registryAuthIssuer struct {
	serviceAccount string
	newTokenSource tokenSourceFactory

	mu    sync.Mutex
	token *oauth2.Token
}

func newRegistryAuthIssuer(cfg *internal.Config) *registryAuthIssuer {
	if cfg == nil || cfg.RunnerContainerImagePullServiceAccount == "" {
		return nil
	}
	return &registryAuthIssuer{
		serviceAccount: cfg.RunnerContainerImagePullServiceAccount,
		newTokenSource: impersonatedTokenSource,
	}
}

func impersonatedTokenSource(ctx context.Context, serviceAccount string) (oauth2.TokenSource, error) {
	return impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: serviceAccount,
		Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
	})
}

// auth returns nil when the image is not hosted on Artifact Registry.
func (i *registryAuthIssuer) auth(imageURL string) (*app.RunnerContainerImageRegistryAuth, error) {
	if i == nil {
		return nil, nil
	}
	registry, ok := artifactRegistryHost(imageURL)
	if !ok {
		return nil, nil
	}

	token, err := i.currentToken()
	if err != nil {
		return nil, err
	}
	return &app.RunnerContainerImageRegistryAuth{
		Registry:  registry,
		Username:  registryAuthUsername,
		Password:  token.AccessToken,
		ExpiresAt: token.Expiry,
	}, nil
}

func (i *registryAuthIssuer) currentToken() (*oauth2.Token, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.token != nil && time.Until(i.token.Expiry) > registryAuthMinRemaining {
		return i.token, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), registryAuthMintTimeout)
	defer cancel()

	ts, err := i.newTokenSource(ctx, i.serviceAccount)
	if err != nil {
		return nil, fmt.Errorf("unable to impersonate %s: %w", i.serviceAccount, err)
	}
	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to mint token for %s: %w", i.serviceAccount, err)
	}
	i.token = token
	return token, nil
}

func artifactRegistryHost(imageURL string) (string, bool) {
	host, _, _ := strings.Cut(imageURL, "/")
	if !strings.HasSuffix(host, "-docker.pkg.dev") {
		return "", false
	}
	return host, true
}

package gar

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
	"github.com/nuonco/nuon/pkg/plugins/configs"
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
		l.Info("getting token using gcloud ADC...")
		token, err := getGARToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get gar token: %w", err)
		}
		l.Info("got token using gcloud ADC")
		username = "oauth2accesstoken"
		password = token
	}

	// GAR login server is derived from the repository URL (e.g., us-docker.pkg.dev)
	loginServer := cfg.LoginServer
	if loginServer == "" {
		loginServer = extractGARLoginServer(cfg.Repository)
	}

	return &registry.AccessInfo{
		Image: cfg.Repository,
		Auth: &registry.AccessInfoAuth{
			Username:      username,
			Password:      password,
			ServerAddress: loginServer,
		},
	}, nil
}

// getGARToken retrieves an access token for GAR using gcloud Application Default Credentials.
func getGARToken(ctx context.Context) (string, error) {
	// First try using gcloud to get an access token
	cmd := exec.CommandContext(ctx, "gcloud", "auth", "print-access-token")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	// Fall back to reading the ADC token file and using it directly
	// This handles cases where gcloud CLI isn't available but ADC credentials are
	cmd = exec.CommandContext(ctx, "gcloud", "auth", "application-default", "print-access-token")
	out, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("unable to get access token from gcloud: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// extractGARLoginServer extracts the GAR login server from a repository URL.
// e.g., "us-docker.pkg.dev/project/repo/image" -> "us-docker.pkg.dev"
func extractGARLoginServer(repo string) string {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) > 0 && strings.Contains(parts[0], "pkg.dev") {
		return parts[0]
	}
	return ""
}

// EncodeDockerAuth creates a base64-encoded Docker auth string.
func EncodeDockerAuth(username, password string) string {
	auth := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

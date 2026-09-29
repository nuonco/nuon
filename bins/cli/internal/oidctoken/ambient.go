package oidctoken

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const (
	tokenEnvVar     = "NUON_OIDC_TOKEN"
	tokenFileEnvVar = "NUON_OIDC_TOKEN_FILE"
	audienceEnvVar  = "NUON_OIDC_AUDIENCE"
)

func Audience(explicit, fallback string) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv(audienceEnvVar); env != "" {
		return env
	}
	return fallback
}

func Available() bool {
	return os.Getenv(tokenEnvVar) != "" ||
		os.Getenv(tokenFileEnvVar) != "" ||
		githubActionsAvailable()
}

func Detect(ctx context.Context, audience string) (token, source string, ok bool, err error) {
	if raw := strings.TrimSpace(os.Getenv(tokenEnvVar)); raw != "" {
		return raw, tokenEnvVar, true, nil
	}

	if path := os.Getenv(tokenFileEnvVar); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", tokenFileEnvVar, true, fmt.Errorf("unable to read %s (%s): %w", tokenFileEnvVar, path, err)
		}
		token := strings.TrimSpace(string(data))
		if token == "" {
			return "", tokenFileEnvVar, true, fmt.Errorf("token file %s is empty", path)
		}
		return token, tokenFileEnvVar, true, nil
	}

	if githubActionsAvailable() {
		token, err := fetchGitHubActionsToken(ctx, audience)
		if err != nil {
			return "", "GitHub Actions", true, err
		}
		return token, "GitHub Actions", true, nil
	}

	return "", "", false, nil
}

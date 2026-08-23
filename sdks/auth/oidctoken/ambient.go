// Package oidctoken discovers ambient OIDC ID tokens in CI so callers can
// exchange them for Nuon API tokens without storing a secret.
//
// It has no dependencies of its own so every Nuon SDK can share one
// implementation. Exchanging what it finds is transport-specific; see package auth.
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

// Audience resolves the audience for ambient token requests: explicit value,
// then NUON_OIDC_AUDIENCE, then fallback. Callers pass their API URL, scoping
// tokens to the target control plane.
func Audience(explicit, fallback string) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv(audienceEnvVar); env != "" {
		return env
	}
	return fallback
}

// Available reports whether an ambient token source is present, without fetching.
func Available() bool {
	return os.Getenv(tokenEnvVar) != "" ||
		os.Getenv(tokenFileEnvVar) != "" ||
		githubActionsAvailable()
}

// Detect returns an ambient OIDC ID token and a label for where it came from.
// Precedence: NUON_OIDC_TOKEN, NUON_OIDC_TOKEN_FILE, GitHub Actions. ok=false
// means no source was configured; an error means a configured source failed.
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

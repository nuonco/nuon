// Package oidctoken discovers ambient OIDC ID tokens in automation
// environments (CI) so the CLI can exchange them for Nuon API tokens without
// any stored secrets.
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

// Audience resolves the audience for ambient token requests: the explicit
// value (flag), then NUON_OIDC_AUDIENCE, then the fallback — callers pass the
// configured API URL so tokens are scoped to the target control plane by
// default.
func Audience(explicit, fallback string) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv(audienceEnvVar); env != "" {
		return env
	}
	return fallback
}

// Available reports whether an ambient OIDC token source is present, without
// fetching anything.
func Available() bool {
	return os.Getenv(tokenEnvVar) != "" ||
		os.Getenv(tokenFileEnvVar) != "" ||
		githubActionsAvailable()
}

// Detect returns an ambient OIDC ID token and a human-readable source label.
// Precedence: NUON_OIDC_TOKEN, NUON_OIDC_TOKEN_FILE, GitHub Actions ID token
// endpoint. Returns ok=false when no source is configured; returns an error
// only when a source is configured but fetching the token fails.
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

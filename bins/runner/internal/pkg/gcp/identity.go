package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/metadata"
)

// GetIdentityToken fetches a GCP identity token (JWT) from the instance metadata service.
// The audience is embedded in the token and must match what the server expects.
func GetIdentityToken(ctx context.Context, audience string) (string, error) {
	token, err := metadata.GetWithContext(ctx, fmt.Sprintf("instance/service-accounts/default/identity?audience=%s&format=full", audience))
	if err != nil {
		return "", fmt.Errorf("failed to get identity token: %w", err)
	}
	return token, nil
}

// GetAccessToken fetches an OAuth2 access token for the default service account.
// The server uses this to independently read instance metadata via the Compute API.
func GetAccessToken(ctx context.Context) (string, error) {
	token, err := metadata.GetWithContext(ctx, "instance/service-accounts/default/token")
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	return token, nil
}

// GetInstanceName returns the instance name from metadata.
func GetInstanceName(ctx context.Context) (string, error) {
	name, err := metadata.InstanceNameWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get instance name: %w", err)
	}
	return name, nil
}

// GetProjectID returns the project ID from metadata.
func GetProjectID(ctx context.Context) (string, error) {
	project, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get project ID: %w", err)
	}
	return project, nil
}

// GetZone returns the instance zone from metadata.
func GetZone(ctx context.Context) (string, error) {
	zone, err := metadata.ZoneWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get zone: %w", err)
	}
	return zone, nil
}

// IsGCPInstance checks if we're running on a GCP instance.
func IsGCPInstance(_ context.Context) bool {
	return metadata.OnGCE()
}

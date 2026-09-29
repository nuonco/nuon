package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/metadata"
)

func GetIdentityToken(ctx context.Context, audience string) (string, error) {
	token, err := metadata.GetWithContext(ctx, fmt.Sprintf("instance/service-accounts/default/identity?audience=%s&format=full", audience))
	if err != nil {
		return "", fmt.Errorf("failed to get identity token: %w", err)
	}
	return token, nil
}

func GetAccessToken(ctx context.Context) (string, error) {
	token, err := metadata.GetWithContext(ctx, "instance/service-accounts/default/token")
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	return token, nil
}

func GetInstanceName(ctx context.Context) (string, error) {
	name, err := metadata.InstanceNameWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get instance name: %w", err)
	}
	return name, nil
}

func GetProjectID(ctx context.Context) (string, error) {
	project, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get project ID: %w", err)
	}
	return project, nil
}

func GetZone(ctx context.Context) (string, error) {
	zone, err := metadata.ZoneWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get zone: %w", err)
	}
	return zone, nil
}

func IsGCPInstance(_ context.Context) bool {
	return metadata.OnGCE()
}

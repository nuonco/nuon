package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/metadata"
)

const runnerIDMetadataKey = "nuon_runner_id"

// GetIdentityToken fetches a GCP identity token (JWT) from the instance metadata service.
// The audience is embedded in the token and must match what the server expects.
func GetIdentityToken(ctx context.Context, audience string) (string, error) {
	token, err := metadata.GetWithContext(ctx, fmt.Sprintf("instance/service-accounts/default/identity?audience=%s&format=full", audience))
	if err != nil {
		return "", fmt.Errorf("failed to get identity token: %w", err)
	}
	return token, nil
}

// GetRunnerID returns the runner ID the stack wrote into instance metadata.
func GetRunnerID(ctx context.Context) (string, error) {
	runnerID, err := metadata.InstanceAttributeValueWithContext(ctx, runnerIDMetadataKey)
	if err != nil {
		return "", fmt.Errorf("failed to get %s from instance metadata: %w", runnerIDMetadataKey, err)
	}
	if runnerID == "" {
		return "", fmt.Errorf("%s instance metadata is empty", runnerIDMetadataKey)
	}
	return runnerID, nil
}

// IsGCPInstance checks if we're running on a GCP instance.
func IsGCPInstance(_ context.Context) bool {
	return metadata.OnGCE()
}

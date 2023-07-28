package gqlclient

import (
	"context"
	"fmt"
)

func (c *client) GetInstanceStatus(ctx context.Context, installID, componentID string, deployID string) (Status, error) {
	installResp, err := getInstall(ctx, c.graphqlClient, installID)
	if err != nil {
		return StatusUnspecified, fmt.Errorf("unable to get install: %w", err)
	}

	statusResp, err := getInstanceStatus(ctx, c.graphqlClient,
		installResp.Install.App.Org.Id,
		installResp.Install.App.Id,
		componentID,
		deployID,
		installID)

	if err != nil {
		return StatusUnspecified, fmt.Errorf("unable to upsert install: %w", err)
	}

	return statusResp.InstanceStatus.Status, nil
}

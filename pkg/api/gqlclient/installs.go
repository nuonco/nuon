package gqlclient

import (
	"context"
	"fmt"
)

// installs
func (c *client) GetInstall(ctx context.Context, installID string) (*getInstallInstall, error) {
	resp, err := getInstall(ctx, c.graphqlClient, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	return &resp.Install, nil
}

func (c *client) GetInstalls(ctx context.Context, appID string) ([]*getInstallsInstallsInstallConnectionEdgesInstallEdgeNodeInstall, error) {
	resp, err := getInstalls(ctx, c.graphqlClient, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get installs: %w", err)
	}

	installs := make([]*getInstallsInstallsInstallConnectionEdgesInstallEdgeNodeInstall, 0)
	for _, install := range resp.Installs.Edges {
		i := install
		installs = append(installs, &i.Node)
	}

	return installs, nil
}

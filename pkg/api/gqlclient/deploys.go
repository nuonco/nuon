package gqlclient

import (
	"context"
	"fmt"
)

type Deploy struct {
	deployFields
}

func (c *client) StartDeploy(ctx context.Context, input DeployInput) (*Deploy, error) {
	startResp, err := startDeploy(ctx, c.graphqlClient, &input)
	if err != nil {
		return nil, fmt.Errorf("unable to start deploy: %w", err)
	}
	return &Deploy{
		startResp.StartDeploy.deployFields,
	}, nil
}

func (c *client) GetDeploy(ctx context.Context, deployID string) (*Deploy, error) {
	deployResp, err := getDeploy(ctx, c.graphqlClient, deployID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to get deploy: %w", err)
	}
	return &Deploy{deployResp.Deploy.deployFields}, nil
}

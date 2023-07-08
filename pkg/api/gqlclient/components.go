package gqlclient

import (
	"context"
	"fmt"
)

// components
func (c *client) GetComponent(ctx context.Context, componentID string) (*getComponentComponent, error) {
	resp, err := getComponent(ctx, c.graphqlClient, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app: %w", err)
	}

	return &resp.Component, nil
}

func (c *client) GetComponents(ctx context.Context, appID string) ([]*getComponentsComponentsComponentConnectionEdgesComponentEdgeNodeComponent, error) {
	resp, err := getComponents(ctx, c.graphqlClient, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get apps: %w", err)
	}

	components := make([]*getComponentsComponentsComponentConnectionEdgesComponentEdgeNodeComponent, 0)
	for _, comp := range resp.Components.Edges {
		c := comp
		components = append(components, &c.Node)
	}

	return components, nil
}

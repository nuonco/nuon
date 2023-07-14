package gqlclient

import (
	"context"
	"fmt"
)

// components
func (c *client) GetComponent(ctx context.Context, componentID string) (*getComponentComponent, error) {
	resp, err := getComponent(ctx, c.graphqlClient, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	return &resp.Component, nil
}

func (c *client) GetComponents(ctx context.Context, appID string) ([]*getComponentsComponentsComponentConnectionEdgesComponentEdgeNodeComponent, error) {
	resp, err := getComponents(ctx, c.graphqlClient, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get components: %w", err)
	}

	components := make([]*getComponentsComponentsComponentConnectionEdgesComponentEdgeNodeComponent, 0)
	for _, comp := range resp.Components.Edges {
		c := comp
		components = append(components, &c.Node)
	}

	return components, nil
}

func (c *client) UpsertComponent(ctx context.Context, input ComponentInput) (*upsertComponentUpsertComponent, error) {
	resp, err := upsertComponent(ctx, c.graphqlClient, input)
	if err != nil {
		return nil, fmt.Errorf("unable to upsertComponent: %w", err)
	}

	return &resp.UpsertComponent, nil
}

func (c *client) DeleteComponent(ctx context.Context, componentID string) error {
	_, err := deleteComponent(ctx, c.graphqlClient, componentID)
	if err != nil {
		return fmt.Errorf("unable to delete component: %w", err)
	}
	return nil
}

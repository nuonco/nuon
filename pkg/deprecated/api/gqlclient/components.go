package gqlclient

import (
	"context"
	"fmt"
)

type Component struct {
	componentFields
}

// components
func (c *client) GetComponent(ctx context.Context, componentID string) (*Component, error) {
	resp, err := getComponent(ctx, c.graphqlClient, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	return &Component{
		resp.Component.componentFields,
	}, nil
}

func (c *client) GetComponents(ctx context.Context, appID string) ([]*Component, error) {
	resp, err := getComponents(ctx, c.graphqlClient, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get components: %w", err)
	}

	components := make([]*Component, 0)
	for _, comp := range resp.Components.Edges {
		components = append(components, &Component{
			comp.Node.componentFields,
		})
	}

	return components, nil
}

func (c *client) UpsertComponent(ctx context.Context, input ComponentInput) (*Component, error) {
	resp, err := upsertComponent(ctx, c.graphqlClient, &input)
	if err != nil {
		return nil, fmt.Errorf("unable to upsertComponent: %w", err)
	}

	return &Component{
		resp.UpsertComponent.componentFields,
	}, nil
}

func (c *client) DeleteComponent(ctx context.Context, componentID string) (bool, error) {
	deleteResp, err := deleteComponent(ctx, c.graphqlClient, componentID)
	if err != nil {
		return false, fmt.Errorf("unable to delete component: %w", err)
	}

	return deleteResp.DeleteComponent, nil
}

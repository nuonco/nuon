package gqlclient

import (
	"context"
	"fmt"
)

func (c *client) GetApp(ctx context.Context, appID string) (*getAppApp, error) {
	resp, err := getApp(ctx, c.graphqlClient, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app: %w", err)
	}

	return &resp.App, nil
}

func (c *client) GetApps(ctx context.Context, orgID string) ([]*getAppsAppsAppConnectionEdgesAppEdgeNodeApp, error) {
	resp, err := getApps(ctx, c.graphqlClient, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get apps: %w", err)
	}

	apps := make([]*getAppsAppsAppConnectionEdgesAppEdgeNodeApp, 0)
	for _, app := range resp.Apps.Edges {
		a := app
		apps = append(apps, &a.Node)
	}

	return apps, nil
}

func (c *client) UpsertApp(ctx context.Context, input AppInput) (*upsertAppUpsertApp, error) {
	resp, err := upsertApp(ctx, c.graphqlClient, input)
	if err != nil {
		return nil, fmt.Errorf("unable to upsertOrg: %w", err)
	}

	return &resp.UpsertApp, nil
}

func (c *client) DeleteApp(ctx context.Context, id string) error {
	_, err := deleteApp(ctx, c.graphqlClient, id)
	if err != nil {
		return fmt.Errorf("unable to upsertOrg: %w", err)
	}

	return nil
}

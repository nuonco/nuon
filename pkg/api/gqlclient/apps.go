package gqlclient

import (
	"context"
	"fmt"
)

type App struct {
	appFields
}

func (c *client) GetApp(ctx context.Context, appID string) (*App, error) {
	resp, err := getApp(ctx, c.graphqlClient, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app: %w", err)
	}

	return &App{
		resp.App.appFields,
	}, nil
}

func (c *client) GetApps(ctx context.Context, orgID string) ([]*App, error) {
	resp, err := getApps(ctx, c.graphqlClient, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get apps: %w", err)
	}

	apps := make([]*App, 0)
	for _, app := range resp.Apps.Edges {
		apps = append(apps, &App{
			app.Node.appFields,
		})
	}

	return apps, nil
}

func (c *client) UpsertApp(ctx context.Context, input AppInput) (*App, error) {
	resp, err := upsertApp(ctx, c.graphqlClient, &input)
	if err != nil {
		return nil, fmt.Errorf("unable to upsertOrg: %w", err)
	}

	return &App{
		resp.UpsertApp.appFields,
	}, nil
}

func (c *client) DeleteApp(ctx context.Context, id string) (bool, error) {
	resp, err := deleteApp(ctx, c.graphqlClient, id)
	if err != nil {
		return false, fmt.Errorf("unable to upsertOrg: %w", err)
	}

	return resp.DeleteApp, nil
}

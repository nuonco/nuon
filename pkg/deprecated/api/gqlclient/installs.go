package gqlclient

import (
	"context"
	"fmt"
)

type Install struct {
	installFields
}

func (c *client) GetInstallStatus(ctx context.Context, orgID, appID, installID string) (Status, error) {
	resp, err := getInstallStatus(ctx, c.graphqlClient, orgID, appID, installID)
	if err != nil {
		return StatusUnspecified, fmt.Errorf("unable to upsert install: %w", err)
	}

	return resp.InstallStatus, nil
}

func (c *client) UpsertInstall(ctx context.Context, input InstallInput) (*Install, error) {
	resp, err := upsertInstall(ctx, c.graphqlClient, &input)
	if err != nil {
		return nil, fmt.Errorf("unable to upsert install: %w", err)
	}

	return &Install{
		resp.UpsertInstall.installFields,
	}, nil
}

func (c *client) DeleteInstall(ctx context.Context, installID string) (bool, error) {
	resp, err := deleteInstall(ctx, c.graphqlClient, installID)
	if err != nil {
		return false, fmt.Errorf("unable to delete install: %w", err)
	}

	return resp.DeleteInstall, nil
}

func (c *client) GetInstall(ctx context.Context, installID string) (*Install, error) {
	resp, err := getInstall(ctx, c.graphqlClient, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	return &Install{
		resp.Install.installFields,
	}, nil
}

func (c *client) GetInstalls(ctx context.Context, appID string) ([]*Install, error) {
	resp, err := getInstalls(ctx, c.graphqlClient, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get installs: %w", err)
	}

	installs := make([]*Install, 0)
	for _, install := range resp.Installs.Edges {
		installs = append(installs, &Install{
			install.Node.installFields,
		})
	}

	return installs, nil
}

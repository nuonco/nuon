package gqlclient

import (
	"context"
	"fmt"
)

func (c *client) GetOrg(ctx context.Context, orgID string) (*getOrgOrg, error) {
	resp, err := getOrg(ctx, c.graphqlClient, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	return &resp.Org, nil
}

func (c *client) GetOrgs(ctx context.Context, userID string) ([]*getOrgsOrgsOrgConnectionEdgesOrgEdgeNodeOrg, error) {
	resp, err := getOrgs(ctx, c.graphqlClient, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to get apps: %w", err)
	}

	orgs := make([]*getOrgsOrgsOrgConnectionEdgesOrgEdgeNodeOrg, 0)
	for _, org := range resp.Orgs.Edges {
		o := org
		orgs = append(orgs, &o.Node)
	}

	return orgs, nil
}

func (c *client) UpsertOrg(ctx context.Context, input OrgInput) (*upsertOrgUpsertOrg, error) {
	resp, err := upsertOrg(ctx, c.graphqlClient, input)
	if err != nil {
		return nil, fmt.Errorf("unable to upsertOrg: %w", err)
	}

	return &resp.UpsertOrg, nil
}

func (c *client) DeleteOrg(ctx context.Context, id string) error {
	_, err := deleteOrg(ctx, c.graphqlClient, id)
	if err != nil {
		return fmt.Errorf("unable to upsertOrg: %w", err)
	}

	return nil
}

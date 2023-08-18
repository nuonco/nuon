package gqlclient

import (
	"context"
	"fmt"
)

type Org struct {
	orgFields
}

func (c *client) GetOrg(ctx context.Context, orgID string) (*Org, error) {
	resp, err := getOrg(ctx, c.graphqlClient, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	return &Org{
		resp.Org.orgFields,
	}, nil
}

func (c *client) GetOrgs(ctx context.Context, userID string) ([]*Org, error) {
	resp, err := getOrgs(ctx, c.graphqlClient, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to get apps: %w", err)
	}

	orgs := make([]*Org, 0)
	for _, org := range resp.Orgs.Edges {
		orgs = append(orgs, &Org{
			org.Node.orgFields,
		})
	}

	return orgs, nil
}

func (c *client) UpsertOrg(ctx context.Context, input OrgInput) (*Org, error) {
	resp, err := upsertOrg(ctx, c.graphqlClient, &input)
	if err != nil {
		return nil, fmt.Errorf("unable to upsertOrg: %w", err)
	}

	return &Org{
		resp.UpsertOrg.orgFields,
	}, nil
}

func (c *client) DeleteOrg(ctx context.Context, id string) error {
	_, err := deleteOrg(ctx, c.graphqlClient, id)
	if err != nil {
		return fmt.Errorf("unable to upsertOrg: %w", err)
	}

	return nil
}

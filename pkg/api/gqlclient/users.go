package gqlclient

import (
	"context"
	"fmt"
)

// users
func (c *client) GetCurrentUser(ctx context.Context) (*getCurrentUserMeUser, error) {
	resp, err := getCurrentUser(ctx, c.graphqlClient)
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	return &resp.Me, nil

}

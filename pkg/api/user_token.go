package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AdminUserRequest struct {
	Email    string `json:"email"`
	Duration string `json:"duration"`
}

type AdminUserResponse struct {
	APIToken string `json:"api_token"`
}

func (c *client) CreateAdminUser(ctx context.Context, email string, duration time.Duration) (*AdminUserResponse, error) {
	endpoint := "/v1/general/admin-user"
	byts, err := c.execPostRequest(ctx, endpoint, AdminUserRequest{
		Email:    email,
		Duration: duration.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute post request: %w", err)
	}

	var response AdminUserResponse
	if err := json.Unmarshal(byts, &response); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &response, nil
}

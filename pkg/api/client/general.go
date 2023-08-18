package client

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/client/client/operations"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

// general methods
func (c *client) GetCurrentUser(ctx context.Context) (*models.AppUserToken, error) {
	resp, err := c.genClient.Operations.GetV1GeneralCurrentUser(&operations.GetV1GeneralCurrentUserParams{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) PublishMetrics(ctx context.Context, req []*models.ServicePublishMetricInput) error {
	_, err := c.genClient.Operations.PostV1GeneralMetrics(&operations.PostV1GeneralMetricsParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return fmt.Errorf("unable to publish metrics: %w", err)
	}

	return nil
}

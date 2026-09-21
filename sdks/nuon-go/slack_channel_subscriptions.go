package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) ListSlackOrgLinks(ctx context.Context) ([]*models.AppSlackOrgLink, error) {
	resp, err := c.genClient.Operations.ListSlackOrgLinks(&operations.ListSlackOrgLinksParams{
		OrgID:   c.OrgID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ListSlackChannelSubscriptions(ctx context.Context) ([]*models.AppSlackChannelSubscription, error) {
	resp, err := c.genClient.Operations.ListSlackChannelSubscriptions(&operations.ListSlackChannelSubscriptionsParams{
		OrgID:   c.OrgID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateSlackChannelSubscription(ctx context.Context, req *models.ServiceCreateChannelSubscriptionRequest) (*models.AppSlackChannelSubscription, error) {
	resp, err := c.genClient.Operations.CreateSlackChannelSubscription(&operations.CreateSlackChannelSubscriptionParams{
		OrgID:   c.OrgID,
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateSlackChannelSubscription(ctx context.Context, subID string, req *models.ServiceUpdateChannelSubscriptionRequest) (*models.AppSlackChannelSubscription, error) {
	resp, err := c.genClient.Operations.UpdateSlackChannelSubscription(&operations.UpdateSlackChannelSubscriptionParams{
		OrgID:   c.OrgID,
		SubID:   subID,
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteSlackChannelSubscription(ctx context.Context, subID string) error {
	_, err := c.genClient.Operations.DeleteSlackChannelSubscription(&operations.DeleteSlackChannelSubscriptionParams{
		OrgID:   c.OrgID,
		SubID:   subID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	return err
}

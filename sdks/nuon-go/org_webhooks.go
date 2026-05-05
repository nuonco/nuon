package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetCurrentOrgWebhooks(ctx context.Context) ([]*models.ServiceCurrentOrgWebhookResponse, error) {
	resp, err := c.genClient.Operations.GetCurrentOrgWebhooks(&operations.GetCurrentOrgWebhooksParams{
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateCurrentOrgWebhook(ctx context.Context, req *models.ServiceCreateCurrentOrgWebhookRequest) (*models.ServiceCurrentOrgWebhookResponse, error) {
	resp, err := c.genClient.Operations.CreateCurrentOrgWebhook(&operations.CreateCurrentOrgWebhookParams{
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateCurrentOrgWebhook(ctx context.Context, webhookID string, req *models.ServiceUpdateCurrentOrgWebhookRequest) (*models.ServiceCurrentOrgWebhookResponse, error) {
	resp, err := c.genClient.Operations.UpdateCurrentOrgWebhook(&operations.UpdateCurrentOrgWebhookParams{
		WebhookID: webhookID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteCurrentOrgWebhook(ctx context.Context, webhookID string) error {
	_, err := c.genClient.Operations.DeleteCurrentOrgWebhook(&operations.DeleteCurrentOrgWebhookParams{
		WebhookID: webhookID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	return err
}

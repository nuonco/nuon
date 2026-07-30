package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// ExchangeOIDCToken exchanges an OIDC ID token (e.g. from GitHub Actions) for
// a short-lived Nuon API token. This call is unauthenticated: trust is
// established by verifying the presented token against the org's OIDC trust
// policies.
func (c *client) ExchangeOIDCToken(ctx context.Context, req *models.ServiceExchangeOIDCTokenRequest) (*models.ServiceExchangeOIDCTokenResponse, error) {
	resp, err := c.genClient.Operations.ExchangeOIDCToken(&operations.ExchangeOIDCTokenParams{
		Context: ctx,
		Req:     req,
	})
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateOIDCTrustPolicy(ctx context.Context, req *models.ServiceCreateOIDCTrustPolicyRequest) (*models.AppOIDCTrustPolicy, error) {
	resp, err := c.genClient.Operations.CreateOIDCTrustPolicy(&operations.CreateOIDCTrustPolicyParams{
		Context: ctx,
		Req:     req,
	}, c.getApiKeyAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ListOIDCTrustPolicies(ctx context.Context) ([]*models.AppOIDCTrustPolicy, error) {
	resp, err := c.genClient.Operations.ListOIDCTrustPolicies(&operations.ListOIDCTrustPoliciesParams{
		Context: ctx,
	}, c.getApiKeyAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetOIDCTrustPolicy(ctx context.Context, policyID string) (*models.AppOIDCTrustPolicy, error) {
	resp, err := c.genClient.Operations.GetOIDCTrustPolicy(&operations.GetOIDCTrustPolicyParams{
		Context:  ctx,
		PolicyID: policyID,
	}, c.getApiKeyAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateOIDCTrustPolicy(ctx context.Context, policyID string, req *models.ServiceUpdateOIDCTrustPolicyRequest) (*models.AppOIDCTrustPolicy, error) {
	resp, err := c.genClient.Operations.UpdateOIDCTrustPolicy(&operations.UpdateOIDCTrustPolicyParams{
		Context:  ctx,
		PolicyID: policyID,
		Req:      req,
	}, c.getApiKeyAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteOIDCTrustPolicy(ctx context.Context, policyID string) error {
	_, err := c.genClient.Operations.DeleteOIDCTrustPolicy(&operations.DeleteOIDCTrustPolicyParams{
		Context:  ctx,
		PolicyID: policyID,
	}, c.getApiKeyAuthInfo())
	return err
}

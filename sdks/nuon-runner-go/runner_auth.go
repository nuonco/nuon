package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) RunnerAuthAWS(ctx context.Context, req *models.GithubComNuoncoNuonServicesCtlAPIInternalAppRunnerAuthServiceRunnerAuthAWSRequest) (*models.GithubComNuoncoNuonServicesCtlAPIInternalAppRunnerAuthServiceRunnerAuthAWSResponse, error) {
	resp, err := c.genClient.Operations.RunnerAuthAWS(&operations.RunnerAuthAWSParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) RunnerAuthGCP(ctx context.Context, req *models.ServiceRunnerAuthGCPRequest) (*models.ServiceRunnerAuthGCPResponse, error) {
	resp, err := c.genClient.Operations.RunnerAuthGCP(&operations.RunnerAuthGCPParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) RunnerAuthAzure(ctx context.Context, req *models.ServiceRunnerAuthAzureRequest) (*models.ServiceRunnerAuthAzureResponse, error) {
	resp, err := c.genClient.Operations.RunnerAuthAzure(&operations.RunnerAuthAzureParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

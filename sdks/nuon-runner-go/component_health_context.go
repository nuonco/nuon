package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) PutComponentHealthContext(ctx context.Context, clusterInfoJSON string, sandboxReleases []string) error {
	_, err := c.genClient.Operations.PutComponentHealthContext(&operations.PutComponentHealthContextParams{
		RunnerID: c.RunnerID,
		Req: &models.ServiceComponentHealthContextRequest{
			ClusterInfoJSON:     &clusterInfoJSON,
			SandboxHelmReleases: sandboxReleases,
		},
		Context: ctx,
	}, c.getAuthInfo())
	return err
}

func (c *client) GetComponentHealthContext(ctx context.Context) (string, []string, error) {
	resp, err := c.genClient.Operations.GetComponentHealthContext(&operations.GetComponentHealthContextParams{
		RunnerID: c.RunnerID,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		return "", nil, err
	}

	return resp.Payload.ClusterInfoJSON, resp.Payload.SandboxHelmReleases, nil
}

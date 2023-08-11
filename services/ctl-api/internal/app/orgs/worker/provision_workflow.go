package worker

import "go.temporal.io/sdk/workflow"

type ProvisionV2Request struct {
	OrgID string
}

type ProvisionV2Response struct{}

func (w *Workflows) ProvisionV2(ctx workflow.Context, req ProvisionV2Request) (ProvisionV2Response, error) {
	return ProvisionV2Response{}, nil
}

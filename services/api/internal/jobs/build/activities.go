package build

import (
	"context"

	tclient "go.temporal.io/sdk/client"
)

type activities struct {
}

func NewActivities(tc tclient.Client) *activities {
	return &activities{}
}

type DoBuildStuffOneResponse struct {
	WorkflowID string
}

func (a *activities) DoBuildStuffOne(ctx context.Context, appID string) (*DoBuildStuffOneResponse, error) {

	return &DoBuildStuffOneResponse{WorkflowID: "iamhardcoded"}, nil
}

package installdelegationdns

import (
	"go.temporal.io/sdk/workflow"
)

type DeprovisionDNSDelegationRequest struct {
	WorkflowID string `json:"workflow_id"`
	Domain     string `json:"domain"`
}

type DeprovisionDNSDelegationResponse struct{}

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{ .CallerID }}-provision-dns-delegation
func (w Wkflow) DeprovisionDNSDelegation(ctx workflow.Context, req *DeprovisionDNSDelegationRequest) (*DeprovisionDNSDelegationResponse, error) {
	// TODO(jm): implement
	return nil, nil
}

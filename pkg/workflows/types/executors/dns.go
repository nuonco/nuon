package executors

import "go.temporal.io/sdk/workflow"

type ProvisionDNSDelegationRequest struct {
	WorkflowID  string   `json:"workflow_id"`
	Domain      string   `json:"domain"`
	Nameservers []string `json:"nameservers"`
}

func (d ProvisionDNSDelegationRequest) Validate() error {
	return nil
}

type ProvisionDNSDelegationResponse struct{}

func ProvisionDNSIDCallback(req *ProvisionDNSDelegationRequest) string {
	return req.WorkflowID
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback ProvisionDNSIDCallback
func ProvisionDNSDelegation(workflow.Context, *ProvisionDNSDelegationRequest) (*ProvisionDNSDelegationResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

type DeprovisionDNSDelegationRequest struct {
	WorkflowID string `json:"workflow_id"`
	Domain     string `json:"domain"`
}

type DeprovisionDNSDelegationResponse struct{}

func DeprovisionDNSIDCallback(req *DeprovisionDNSDelegationRequest) string {
	return req.WorkflowID
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback DeprovisionDNSIDCallback
func DeprovisionDNSDelegation(workflow.Context, *DeprovisionDNSDelegationRequest) (*DeprovisionDNSDelegationResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

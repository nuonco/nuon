package executors

import "go.temporal.io/sdk/workflow"

const (
	ProvisionIAMWorkflowName   string = "ProvisionIAM"
	DeprovisionIAMWorkflowName string = "DeprovisionIAM"
)

func ProvisionIDCallback(req *ProvisionIAMRequest) string {
	return "provision-iam-" + req.OrgID
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback ProvisionIDCallback
func ProvisionIAM(workflow.Context, *ProvisionIAMRequest) (*ProvisionIAMResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

type ProvisionIAMRequest struct {
	OrgID       string `json:"org_id"`
	Reprovision bool   `json:"reprovision"`
}

type ProvisionIAMResponse struct {
	RunnerRoleArn string
}

type DeprovisionIAMRequest struct {
	OrgID string
}

type DeprovisionIAMResponse struct{}

func DeprovisionIDCallback(req *DeprovisionIAMRequest) string {
	return "deprovision-iam-" + req.OrgID
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback DeprovisionIDCallback
func DeprovisionIAM(workflow.Context, *DeprovisionIAMRequest) (*DeprovisionIAMResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

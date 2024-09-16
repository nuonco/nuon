package executors

import "go.temporal.io/sdk/workflow"

const (
	ProvisionIAMWorkflowName   string = "ProvisionIAM"
	DeprovisionIAMWorkflowName string = "DeprovisionIAM"
)

// @disabled-temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
func ProvisionIAM(workflow.Context, *ProvisionIAMRequest) (*ProvisionIAMResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

type ProvisionIAMRequest struct {
	OrgID       string `json:"org_id"`
	Reprovision bool   `json:"reprovision"`
}

type ProvisionIAMResponse struct {
	DeploymentsRoleArn   string
	InstallationsRoleArn string
	OdrRoleArn           string
	InstancesRoleArn     string
	InstallerRoleArn     string
	OrgsRoleArn          string
	SecretsRoleArn       string
}

type DeprovisionIAMRequest struct {
	OrgID string
}

type DeprovisionIAMResponse struct{}

// @disabled-temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
func DeprovisionIAM(workflow.Context, *DeprovisionIAMRequest) (*DeprovisionIAMResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

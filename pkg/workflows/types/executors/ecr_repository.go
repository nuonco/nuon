package executors

import (
	"go.temporal.io/sdk/workflow"
)

const (
	ProvisionECRRepositoryWorkflowName   string = "ProvisionECRRepository"
	DeprovisionECRRepositoryWorkflowName string = "DeprovisionECRRepository"
)

type ProvisionECRRepositoryRequest struct {
	OrgID string
	AppID string
}

type ProvisionECRRepositoryResponse struct {
	RegistryID     string
	RepositoryName string
	RepositoryARN  string
	RepositoryURI  string
}

func ProvisionECRRepositoryIDCallback(req *ProvisionECRRepositoryRequest) string {
	return "provision-ecr-" + req.OrgID + "-" + req.AppID
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback ProvisionECRRepositoryIDCallback
func ProvisionECRRepository(workflow.Context, *ProvisionECRRepositoryRequest) (*ProvisionECRRepositoryResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

type DeprovisionECRRepositoryRequest struct {
	OrgID string
	AppID string
}

type DeprovisionECRRepositoryResponse struct{}

func DeprovisionECRRepositoryIDCallback(req *DeprovisionECRRepositoryRequest) string {
	return "deprovision-ecr-" + req.OrgID + "-" + req.AppID
}

// @disabled-temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback DeprovisionECRRepositoryIDCallback
func DeprovisionECRRepository(workflow.Context, *DeprovisionECRRepositoryRequest) (*DeprovisionECRRepositoryResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

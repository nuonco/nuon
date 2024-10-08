package executors

import (
	"go.temporal.io/sdk/workflow"
)

const (
	ProvisionRunnerWorkflowName   = "ProvisionRunner"
	DeprovisionRunnerWorkflowName = "DeprovisionRunner"
)

type ProvisionRunnerRequestImage struct {
	URL string `validate:"required"`
	Tag string `validate:"tag"`
}

type ProvisionRunnerRequest struct {
	RunnerID string                      `validate:"required"`
	APIURL   string                      `validate:"required"`
	APIToken string                      `validate:"required"`
	Image    ProvisionRunnerRequestImage `validate:"required"`
}

func ProvisionRunnerIDCallback(req *ProvisionRunnerRequest) string {
	return "provision-runner-" + req.RunnerID
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback ProvisionRunnerIDCallback
func ProvisionRunner(workflow.Context, *ProvisionRunnerRequest) (*ProvisionRunnerResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

type ProvisionRunnerResponse struct{}

type DeprovisionRunnerRequest struct {
	RunnerID string `validate:"required"`
}

type DeprovisionRunnerResponse struct{}

func DeprovisionRunnerIDCallback(req *DeprovisionRunnerRequest) string {
	return "deprovision-runner-" + req.RunnerID
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback DeprovisionRunnerIDCallback
func DeprovisionRunner(workflow.Context, *DeprovisionRunnerRequest) (*DeprovisionRunnerResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}

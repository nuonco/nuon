package orgs

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	DeprovisionWorkflowName = "Deprovision"
)

type DeprovisionRequest struct {
	OrgId  string `json:"org_id"`
	Region string `json:"region"`
}

type DeprovisionResponse struct{}

func DeprovisionIDCallback(req *DeprovisionRequest) string {
	return fmt.Sprintf("%s-deprovision", req.OrgId)
}

// @temporal-gen workflow
// @execution-timeout 20m
// @task-timeout 10m
// @task-queue "default"
// @id-callback DeprovisionIDCallback
func Deprovision(workflow.Context, *DeprovisionRequest) (*DeprovisionResponse, error) {
	panic("stub for code generation")
}

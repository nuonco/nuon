package canary

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	DeprovisionWorkflowName = "Deprovision"
)

type DeprovisionRequest struct {
	CanaryId    string `json:"canary_id"`
	SandboxMode bool   `json:"sandbox_mode"`
}

type DeprovisionResponse struct {
	CanaryId string `json:"canary_id"`
}

func DeprovisionIDCallback(req *DeprovisionRequest) string {
	return fmt.Sprintf("%s-deprovision", req.CanaryId)
}

// @temporal-gen workflow
// @execution-timeout 24h
// @task-timeout 1h
// @task-queue "default"
// @namespace "canary"
// @id-callback DeprovisionIDCallback
func Deprovision(workflow.Context, *DeprovisionRequest) (*DeprovisionResponse, error) {
	panic("stub for code generation")
}

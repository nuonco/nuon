package canary

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	ProvisionWorkflowName = "Provision"
)

type ProvisionRequest struct {
	CanaryId    string `json:"canary_id"`
	SandboxMode bool   `json:"sandbox_mode"`
}

func (r *ProvisionRequest) Validate() error {
	if r.CanaryId != "" && len(r.CanaryId) != 26 {
		return fmt.Errorf("canary_id must be 26 characters when provided")
	}
	return nil
}

type ProvisionResponse struct {
	CanaryId string `json:"canary_id"`
	OrgId    string `json:"org_id"`
}

func ProvisionIDCallback(req *ProvisionRequest) string {
	return fmt.Sprintf("%s-provision", req.CanaryId)
}

// @temporal-gen workflow
// @execution-timeout 24h
// @task-timeout 1h
// @task-queue "default"
// @namespace "canary"
// @id-callback ProvisionIDCallback
func Provision(workflow.Context, *ProvisionRequest) (*ProvisionResponse, error) {
	panic("stub for code generation")
}

package orgs

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	ProvisionWorkflowName = "Provision"
)

type ProvisionRequest struct {
	OrgId       string `json:"org_id"`
	Region      string `json:"region"`
	Reprovision bool   `json:"reprovision"`
	CustomCert  bool   `json:"custom_cert"`
}

type ProvisionResponse struct{}

func ProvisionIDCallback(req *ProvisionRequest) string {
	return fmt.Sprintf("%s-provision", req.OrgId)
}

// @temporal-gen workflow
// @execution-timeout 20m
// @task-timeout 10m
// @task-queue "default"
// @id-callback ProvisionIDCallback
func Provision(workflow.Context, *ProvisionRequest) (*ProvisionResponse, error) {
	panic("stub for code generation")
}

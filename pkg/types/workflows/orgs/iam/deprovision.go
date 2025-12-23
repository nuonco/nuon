package iam

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	DeprovisionIAMWorkflowName = "DeprovisionIAM"
)

type DeprovisionIAMRequest struct {
	OrgId string `json:"org_id"`
}

type DeprovisionIAMResponse struct {
	Status map[string]interface{} `json:"status"`
}

func DeprovisionIAMIDCallback(req *DeprovisionIAMRequest) string {
	return fmt.Sprintf("%s-deprovision-iam", req.OrgId)
}

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 10m
// @task-queue "default"
// @id-callback DeprovisionIAMIDCallback
func DeprovisionIAM(workflow.Context, *DeprovisionIAMRequest) (*DeprovisionIAMResponse, error) {
	panic("stub for code generation")
}

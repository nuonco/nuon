package plan

import (
	"go.temporal.io/sdk/workflow"
)

const (
	CreatePlanWorkflowName = "CreatePlan"
)

// CreatePlanRequest is a child workflow request for the plan phase.
// The Input field should contain the specific input type (component, sandbox, or runner).
type CreatePlanRequest struct {
	Input interface{} `json:"input"`
}

// PlanRef contains reference information for a plan.
type PlanRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Plan contains the plan details.
type Plan struct {
	ID       string                 `json:"id"`
	Status   string                 `json:"status"`
	Actions  []PlanAction           `json:"actions"`
	Metadata map[string]interface{} `json:"metadata"`
}

// PlanAction represents a single action in a plan.
type PlanAction struct {
	Type     string                 `json:"type"`
	Resource string                 `json:"resource"`
	Action   string                 `json:"action"`
	Details  map[string]interface{} `json:"details"`
}

// CreatePlanResponse is a child workflow response for the plan phase.
type CreatePlanResponse struct {
	Ref  *PlanRef `json:"ref"`
	Plan *Plan    `json:"plan"`
}

// FakePlanResponse returns a fake response for sandbox mode.
func FakePlanResponse() *CreatePlanResponse {
	return &CreatePlanResponse{
		Ref: &PlanRef{
			ID:   "fake-plan-id",
			Type: "sandbox",
		},
		Plan: &Plan{
			ID:       "fake-plan-id",
			Status:   "completed",
			Actions:  []PlanAction{},
			Metadata: map[string]interface{}{},
		},
	}
}

func CreatePlanIDCallback(req *CreatePlanRequest) string {
	return "create-plan"
}

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
// @task-queue "executors"
// @id-callback CreatePlanIDCallback
func CreatePlan(workflow.Context, *CreatePlanRequest) (*CreatePlanResponse, error) {
	panic("stub for code generation")
}

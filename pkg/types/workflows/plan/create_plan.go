package plan

import (
	"go.temporal.io/sdk/workflow"
)

const (
	CreatePlanWorkflowName = "CreatePlan"
)

type CreatePlanRequest struct {
	Input interface{} `json:"input"`
}

type PlanRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type Plan struct {
	ID       string                 `json:"id"`
	Status   string                 `json:"status"`
	Actions  []PlanAction           `json:"actions"`
	Metadata map[string]interface{} `json:"metadata"`
}

type PlanAction struct {
	Type     string                 `json:"type"`
	Resource string                 `json:"resource"`
	Action   string                 `json:"action"`
	Details  map[string]interface{} `json:"details"`
}

type CreatePlanResponse struct {
	Ref  *PlanRef `json:"ref"`
	Plan *Plan    `json:"plan"`
}

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

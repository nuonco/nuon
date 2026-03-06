package examples

import (
	"go.temporal.io/sdk/workflow"
)

type IDTemplateInput struct {
	OrgID  string
	UserID string
}

type TaskInput struct {
	TaskID string
}

// WorkflowWithStructIDTemplate demonstrates ID template with struct input
// @temporal-gen-v2 workflow
// @id-template org-{{.Req.OrgID}}-user-{{.Req.UserID}}
func WorkflowWithStructIDTemplate(ctx workflow.Context, input IDTemplateInput) error {
	return nil
}

// WorkflowWithCallerIDTemplate demonstrates nested workflow IDs using CallerID
// @temporal-gen-v2 workflow
// @id-template {{.CallerID}}-subtask-{{.Req.TaskID}}
func WorkflowWithCallerIDTemplate(ctx workflow.Context, input TaskInput) error {
	return nil
}

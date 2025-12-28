package examples

import (
	"go.temporal.io/sdk/workflow"
)

type IDTemplateInput struct {
	OrgID  string
	UserID string
}

// WorkflowWithStructIDTemplate demonstrates ID template with struct input
// @temporal-gen-v2 workflow
// @id-template org-{{.OrgID}}-user-{{.UserID}}
func WorkflowWithStructIDTemplate(ctx workflow.Context, input IDTemplateInput) error {
	return nil
}

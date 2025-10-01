package workflow

import "fmt"

// NOTE(jm): this should not be required once temporal-gen supports passing in a workflow-id via parameter
func WorkflowIDCallback(req *GenerateWorkflowStepsRequest) string {
	return fmt.Sprintf("generate-steps-%s", req.WorkflowID)
}

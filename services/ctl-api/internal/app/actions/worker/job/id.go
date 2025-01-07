package job

import "fmt"

// NOTE(jm): this should not be required once temporal-gen supports passing in a workflow-id via parameter
func WorkflowIDCallback(req *ExecuteJobRequest) string {
	if req.WorkflowID != "" {
		return req.WorkflowID
	}

	return fmt.Sprintf("execute-job-%s", req.WorkflowID)
}

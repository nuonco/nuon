package job

import "fmt"

func WorkflowIDCallback(req *ExecuteJobRequest) string {
	if req.WorkflowID != "" {
		return req.WorkflowID
	}

	return fmt.Sprintf("execute-job-%s", req.JobID)
}

package app

// WorkflowResponse is returned by endpoints that trigger a workflow
// and have no other meaningful return data.
type WorkflowResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// EmptyResponse is a structured replacement for bare "ok" string or bool responses.
// It represents a successful operation with no meaningful return data.
type EmptyResponse struct{}

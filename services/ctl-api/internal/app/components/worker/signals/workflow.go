package signals

import "fmt"

const (
	EventLoopWorkflowName string = "ComponentEventLoop"
)

type ComponentEventLoopRequest struct {
	ComponentID string
	SandboxMode bool
}

func EventLoopWorkflowID(componentID string) string {
	return fmt.Sprintf("%s-event-loop", componentID)
}

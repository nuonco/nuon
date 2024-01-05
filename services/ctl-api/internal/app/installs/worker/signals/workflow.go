package signals

import "fmt"

const (
	EventLoopWorkflowName string = "InstallEventLoop"
)

func EventLoopWorkflowID(installID string) string {
	return fmt.Sprintf("%s-event-loop", installID)
}

type InstallEventLoopRequest struct {
	InstallID   string
	SandboxMode bool
}

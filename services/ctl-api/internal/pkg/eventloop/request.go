package eventloop

type EventLoopRequest struct {
	ID          string
	SandboxMode bool

	// state managed between different signals
	Version                        string
	RestartCount       int
	VersionChangeCount int
}

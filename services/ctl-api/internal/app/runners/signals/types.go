package signals

// TemporalNamespace is the Temporal namespace used for runner workflows.
const TemporalNamespace string = "runners"

// RequestSignal is the parameter type for runner workflow functions.
// It replaces the legacy eventloop-based RequestSignal.
type RequestSignal struct {
	ID          string
	SandboxMode bool

	Type                     string
	JobID                    string
	HealthCheckID            string
	InstallStackVersionRunID string
	ForceDelete              bool
}

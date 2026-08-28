package customermanaged

import "time"

const (
	RunnerHeartbeatKey                     = "heartbeat.json"
	LegacyRunnerHeartbeatKey               = "runner/heartbeat.json"
	RunnerCapabilityCandidateArtifactPlans = "candidate-artifact-plans-v1"
)

type RunnerHeartbeat struct {
	RunnerID     string    `json:"runner_id"`
	SessionID    string    `json:"session_id"`
	Version      string    `json:"version"`
	BundleDigest string    `json:"bundle_digest"`
	Capabilities []string  `json:"capabilities,omitempty"`
	StartedAt    time.Time `json:"started_at"`
	ObservedAt   time.Time `json:"observed_at"`
}

func (h RunnerHeartbeat) Supports(capability string) bool {
	for _, supported := range h.Capabilities {
		if supported == capability {
			return true
		}
	}
	return false
}

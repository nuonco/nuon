package helpers

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

// PolicyEvaluationContext contains context information for policy evaluation.
// This unified struct is used both during policy preparation (in activities)
// and for reporting results (in workflow conductors).
type PolicyEvaluationContext struct {
	// Core identifiers
	OrgID            string
	AppID            string
	InstallID        *string
	InstallSandboxID *string
	ComponentID      *string
	ComponentBuildID *string

	// Policy evaluation metadata
	PolicyIDs  []string
	InputCount int

	// Human-readable names for display in reports
	OrgName       string
	AppName       string
	InstallName   string
	ComponentName string

	// Fields used during policy preparation/resolution
	AppConfigID   string
	ComponentType app.ComponentType
	IsSandbox     bool
}

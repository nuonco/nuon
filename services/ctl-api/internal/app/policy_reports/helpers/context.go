package helpers

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

type PolicyEvaluationContext struct {
	OrgID            string
	AppID            string
	InstallID        *string
	InstallSandboxID *string
	ComponentID      *string
	ComponentBuildID *string

	PolicyIDs  []string
	InputCount int

	OrgName       string
	AppName       string
	InstallName   string
	ComponentName string

	AppConfigID   string
	ComponentType app.ComponentType
	SandboxType   string
	IsSandbox     bool
}

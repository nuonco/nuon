package state

func NewInstallState() *InstallState {
	return &InstallState{
		Org: orgState{},
		App: appState{
			Secrets: make(map[string]string, 0),
		},
		Install: installState{
			Sandbox: installSandboxState{
				Outputs: make(map[string]interface{}),
			},
			Inputs: make(map[string]string),
		},
		Components: make(map[string]*ComponentState),
	}
}

type installSandboxState struct {
	Type    string                 `json:"type"`
	Version string                 `json:"version"`
	Outputs map[string]interface{} `json:"outputs" faker:"-"`
}

type installState struct {
	ID string `json:"id"`

	PublicDomain   string              `json:"public_domain"`
	InternalDomain string              `json:"internal_domain"`
	Sandbox        installSandboxState `json:"sandbox"`
	Inputs         map[string]string   `json:"inputs"`
}

// InstallState represents the current state of the install.
// Copied from the workers-executors variables backend.
type InstallState struct {
	Org        orgState                   `json:"org"`
	App        appState                   `json:"app"`
	Install    installState               `json:"install"`
	Components map[string]*ComponentState `json:"components"`
}

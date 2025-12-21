package state

func NewInstallState() *InstallState {
	return &InstallState{
		Inputs: make(map[string]string, 0),
	}
}

// TODO(jm): this is going to be deprecated
type InstallState struct {
	Populated bool   `json:"populated"`
	ID        string `json:"id"`
	Name      string `json:"name"`

	PublicDomain   string            `json:"public_domain"`
	InternalDomain string            `json:"internal_domain"`
	Sandbox        SandboxState      `json:"sandbox"`
	Inputs         map[string]string `json:"inputs"`
}

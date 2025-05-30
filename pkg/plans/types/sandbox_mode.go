package plantypes

type MinSandboxMode struct {
	SandboxMode *SandboxMode `json:"omitzero,omitempty"`
}

type TerraformSandboxMode struct {
	// needs to be the outputs of `terraform show -json`
	StateJSON   []byte `json:"state_json"`
	WorkspaceID string `json:"workspace_id"`
}

type HelmSandboxMode struct {
	// write resources into the api
}

type SandboxMode struct {
	Enabled bool `json:"enabled"`

	Outputs map[string]any `json:"outputs"`

	*TerraformSandboxMode `json:"terraform,omitzero,omitempty"`
	*HelmSandboxMode      `json:"helm,omitzero,omitempty"`
}

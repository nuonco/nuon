package plantypes

type MinSandboxMode struct {
	SandboxMode *SandboxMode `json:"sandbox_mode,omitzero,omitempty"`
}

type TerraformSandboxMode struct {
	// needs to be the outputs of `terraform show -json`
	StateJSON   []byte `json:"state_json"`
	WorkspaceID string `json:"workspace_id"`

	// create the plan output
	PlanContents        string `json:"plan_contents"`
	PlanDisplayContents string `json:"plan_display_contents"`
}

type HelmSandboxMode struct {
	PlanContents        string `json:"plan_contents"`
	PlanDisplayContents string `json:"plan_display_contents"`
}

type KubernetesSandboxMode struct {
	PlanContents        string `json:"plan_contents"`
	PlanDisplayContents string `json:"plan_display_contents"`
}

type SandboxMode struct {
	Enabled bool `json:"enabled"`

	Outputs map[string]any `json:"outputs"`

	Terraform          *TerraformSandboxMode  `json:"terraform,omitzero,omitempty"`
	Helm               *HelmSandboxMode       `json:"helm,omitzero,omitempty"`
	KubernetesManifest *KubernetesSandboxMode `json:"kubernetes_manifest,omitzero,omitempty"`
}

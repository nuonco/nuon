package plantypes

type MinSandboxMode struct {
	SandboxMode *SandboxMode `json:"sandbox_mode,omitzero,omitempty"`
}

// fakeSandboxDigest must be a real sha256 (64 lowercase hex chars): sandbox
// refs are parsed as container references, not just interpolated into
// templates, so a decorative placeholder fails to parse.
const fakeSandboxDigest = "sha256:a123b456c789d012e345f678a901b234c567d890a123b456c789d012e345f678"

// FakeOCISyncOutputs is the outputs map a sandboxed oci-sync job reports. It
// mirrors the live runner sync output shape: `repository` and `tag` are the
// bare repo and resolved tag that user templates compose as
// `{{.repository}}:{{.tag}}`, `ref` is the additive digest-pinned form, and
// `display_tag` carries the human-friendly tag.
func FakeOCISyncOutputs(repository, tag string) map[string]any {
	return map[string]any{
		"image": map[string]any{
			"repository":    repository,
			"tag":           tag,
			"ref":           repository + "@" + fakeSandboxDigest,
			"display_tag":   tag,
			"media_type":    "application/vnd.docker.distribution.manifest.v2+json",
			"digest":        fakeSandboxDigest,
			"size":          28437192,
			"urls":          []string{repository + ":" + tag},
			"annotations":   map[string]string{"org.opencontainers.image.created": "2024-04-29T10:15:30Z"},
			"artifact_type": "application/vnd.docker.container.image.v1+json",
			"platform": map[string]any{
				"architecture": "arm64",
				"os":           "linux",
				"os_version":   "10.0",
			},
		},
	}
}

type TerraformSandboxMode struct {
	// needs to be the outputs of `terraform show -json`
	StateJSON   []byte `json:"state_json" swaggertype:"primitive,string"`
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

type PulumiSandboxMode struct {
	WorkspaceID         string `json:"workspace_id"`
	PlanContents        string `json:"plan_contents"`
	PlanDisplayContents string `json:"plan_display_contents"`
}

type SandboxMode struct {
	Enabled bool `json:"enabled"`

	Outputs map[string]any `json:"outputs"`

	Terraform          *TerraformSandboxMode  `json:"terraform,omitzero,omitempty"`
	Helm               *HelmSandboxMode       `json:"helm,omitzero,omitempty"`
	KubernetesManifest *KubernetesSandboxMode `json:"kubernetes_manifest,omitzero,omitempty"`
	Pulumi             *PulumiSandboxMode     `json:"pulumi,omitzero,omitempty"`
}

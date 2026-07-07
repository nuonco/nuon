package terraform

import (
	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// gcpAdapter binds the Terraform engine to the install-stacks/gcp module.
type gcpAdapter struct{}

var _ moduleAdapter = gcpAdapter{}

func (gcpAdapter) ModuleSubdir() string { return "gcp" }

func (gcpAdapter) RenderTFVars(cfg *core.Config) ([]byte, error) { return renderGCPTFVars(cfg) }

func (gcpAdapter) MapOutputs(meta map[string]tfexec.OutputMeta) (*core.Outputs, error) {
	return gcpOutputsToCore(meta)
}

func (gcpAdapter) BackendType() string { return "gcs" }

func (gcpAdapter) BackendConfigKV(be *core.TerraformBackend) []string {
	return []string{
		"bucket=" + be.Bucket,
		"prefix=" + be.Prefix,
	}
}

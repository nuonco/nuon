package terraform

import (
	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// awsAdapter binds the Terraform engine to the install-stacks/aws module.
// renderTFVars and outputsToCore live in tfvars.go / outputs.go.
type awsAdapter struct{}

var _ moduleAdapter = awsAdapter{}

func (awsAdapter) ModuleSubdir() string { return "aws" }

func (awsAdapter) RenderTFVars(cfg *core.Config) ([]byte, error) { return renderTFVars(cfg) }

func (awsAdapter) MapOutputs(meta map[string]tfexec.OutputMeta) (*core.Outputs, error) {
	return outputsToCore(meta)
}

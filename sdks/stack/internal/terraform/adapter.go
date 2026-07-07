package terraform

import (
	"fmt"

	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// moduleAdapter is the per-cloud half of the Terraform method. The engine
// (binary download, module fetch, init/apply/destroy, output reading) is
// shared across clouds; the adapter supplies the three things that differ by
// cloud: which install-stacks subdir holds the module, how the install Config
// maps to that module's tfvars, and how its outputs map back into core.Outputs.
type moduleAdapter interface {
	// ModuleSubdir is the install-stacks subdirectory holding the cloud's
	// module (e.g. "aws", "gcp"). A non-empty Config.TerraformModuleSubdir
	// overrides this.
	ModuleSubdir() string
	// RenderTFVars produces the terraform.tfvars.json bytes for the cloud's
	// module from the install Config.
	RenderTFVars(cfg *core.Config) ([]byte, error)
	// MapOutputs translates the module's `terraform output` result into the
	// cloud-agnostic core.Outputs.
	MapOutputs(meta map[string]tfexec.OutputMeta) (*core.Outputs, error)

	// BackendType is the terraform remote-state backend name for the cloud
	// ("s3" for AWS, "gcs" for GCP).
	BackendType() string
	// BackendConfigKV returns the `key=value` pairs configuring the backend,
	// passed to `terraform init -backend-config`.
	BackendConfigKV(be *core.TerraformBackend) []string
}

// adapterFor returns the module adapter for the given cloud.
func adapterFor(cloud core.Cloud) (moduleAdapter, error) {
	switch cloud {
	case core.CloudAWS:
		return awsAdapter{}, nil
	case core.CloudGCP:
		return gcpAdapter{}, nil
	default:
		return nil, fmt.Errorf("terraform method: no module adapter for cloud %q", cloud)
	}
}

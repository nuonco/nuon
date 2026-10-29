package terraform

import (
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/workspace"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

const (
	defaultFileType string = "file/terraform"
)

type handlerState struct {
	// set during the fetch/validate phase
	plan *plantypes.BuildPlan
	cfg  *plantypes.TerraformBuildPlan

	// fields set by the plugin execution
	workspace      workspace.Workspace
	arch           ociarchive.Archive
	source         *sourcefs.FS
	archiveSource  *sourcefs.FS
	resultTag      string
	jobExecutionID string
	jobID          string
	regCfg         *configs.OCIRegistryRepository
}

package pulumi

import (
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/workspace"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

const (
	defaultFileType string = "file/pulumi"
)

type handlerState struct {
	plan *plantypes.BuildPlan
	cfg  *plantypes.PulumiBuildPlan

	workspace      workspace.Workspace
	arch           ociarchive.Archive
	source         *sourcefs.FS
	resultTag      string
	jobExecutionID string
	jobID          string
	regCfg         *configs.OCIRegistryRepository
}

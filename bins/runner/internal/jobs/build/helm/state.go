package helm

import (
	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/workspace"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

const (
	defaultFileType             string = "file/helm"
	defaultChartPackageFilename string = "chart.tgz"
)

type handlerState struct {
	// set during the fetch/validate phase
	plan *plantypes.BuildPlan
	cfg  *plantypes.HelmBuildPlan

	// fields set by the plugin execution
	workspace      workspace.Workspace
	arch           ociarchive.Archive
	resultTag      string
	jobExecutionID string
	jobID          string
	regCfg         *configs.OCIRegistryRepository

	packagePath string
}

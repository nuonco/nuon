package kubernetes_manifest

import (
	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

const (
	defaultManifestFilename string = "manifest.yaml"
	defaultFileType         string = "application/x-yaml"
)

type handlerState struct {
	plan *plantypes.BuildPlan
	cfg  *plantypes.KubernetesManifestBuildPlan

	workspace      workspace.Workspace
	arch           ociarchive.Archive
	resultTag      string
	jobExecutionID string
	jobID          string
	regCfg         *configs.OCIRegistryRepository
}

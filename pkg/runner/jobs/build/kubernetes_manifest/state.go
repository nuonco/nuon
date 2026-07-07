package kubernetes_manifest

import (
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/workspace"
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

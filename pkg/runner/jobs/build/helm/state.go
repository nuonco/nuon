package helm

import (
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/workspace"
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
	chartPath   string

	policyInput []AdmissionReviewInput
}

type AdmissionReviewInput struct {
	Review AdmissionReviewRequest `json:"review"`
}

type AdmissionReviewRequest struct {
	Kind   AdmissionReviewKind    `json:"kind"`
	Object map[string]interface{} `json:"object"`
}

type AdmissionReviewKind struct {
	Kind    string `json:"kind"`
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
}

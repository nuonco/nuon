package terraform

import (
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/workspace"
	terraformworkspace "github.com/nuonco/nuon/pkg/terraform/workspace"
)

const (
	defaultFileType string = "file/terraform"
)

type handlerState struct {
	workspace workspace.Workspace

	// ociArch is set when the plan uses OCI source instead of git.
	// The unpacked archive directory is used as the source path.
	ociArch ociarchive.Archive

	timeout time.Duration

	// fields set by the plugin execution
	jobExecutionID string
	jobID          string
	tfWorkspace    terraformworkspace.Workspace

	plan       *plantypes.SandboxRunPlan
	appCfg     *models.AppAppConfig
	sandboxCfg *models.AppAppSandboxConfig

	auth *pkgplantypes.PlanAuth
}

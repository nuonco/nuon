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

	ociArch ociarchive.Archive

	timeout time.Duration

	jobExecutionID string
	jobID          string
	tfWorkspace    terraformworkspace.Workspace

	plan       *plantypes.SandboxRunPlan
	appCfg     *models.AppAppConfig
	sandboxCfg *models.AppAppSandboxConfig

	auth *pkgplantypes.PlanAuth
}

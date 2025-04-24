package terraform

import (
	"time"

	"github.com/nuonco/nuon-runner-go/models"

	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	terraformworkspace "github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

const (
	defaultFileType string = "file/terraform"
)

type handlerState struct {
	// set during the fetch/validate phase
	plan         *plantypes.DeployPlan
	appCfg       *models.AppAppConfig
	terraformCfg *models.AppTerraformModuleComponentConfig

	srcCfg  *configs.OCIRegistryRepository
	srcTag  string
	timeout time.Duration

	// fields set by the plugin execution
	arch           ociarchive.Archive
	jobExecutionID string
	jobID          string
	tfWorkspace    terraformworkspace.Workspace
}

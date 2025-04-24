package helm

import (
	"time"

	"github.com/nuonco/nuon-runner-go/models"

	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/pkg/kube"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

const (
	defaultFileType string = "file/helm"
)

type handlerState struct {
	// set during the fetch/validate phase
	plan    *plantypes.DeployPlan
	appCfg  *models.AppAppConfig
	helmCfg *models.AppHelmComponentConfig

	srcCfg  *configs.OCIRegistryRepository
	srcTag  string
	timeout time.Duration

	// fields set by the plugin execution
	arch           ociarchive.Archive
	chartPath      string
	jobExecutionID string
	jobID          string
	clusterInfo    *kube.ClusterInfo
	outputs        map[string]interface{}
}

package helm

import (
	"time"

	"github.com/nuonco/nuon-runner-go/models"

	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
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

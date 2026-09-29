package helm

import (
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

const (
	defaultFileType string = "file/helm"
)

type handlerState struct {
	plan    *plantypes.DeployPlan
	appCfg  *models.AppAppConfig
	helmCfg *models.AppHelmComponentConfig

	auth *pkgplantypes.PlanAuth

	srcCfg  *configs.OCIRegistryRepository
	srcTag  string
	timeout time.Duration

	arch           ociarchive.Archive
	chartPath      string
	jobExecutionID string
	jobID          string
	clusterInfo    *kube.ClusterInfo
	outputs        map[string]interface{}
}

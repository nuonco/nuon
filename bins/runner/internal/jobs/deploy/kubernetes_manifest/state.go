package kubernetes_manifest

import (
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

type handlerState struct {
	plan                              *plantypes.DeployPlan
	appCfg                            *models.AppAppConfig
	kubernetesManifestComponentConfig *models.AppKubernetesManifestComponentConfig
	previousDeployResources           *string

	jobExecutionID string
	jobID          string
	timeout        time.Duration

	outputs map[string]interface{}

	kubeClient *kubernetesClient

	auth *pkgplantypes.PlanAuth

	arch   ociarchive.Archive
	srcCfg *configs.OCIRegistryRepository
	srcTag string
}

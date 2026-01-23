package kubernetes_manifest

import (
	"time"

	"github.com/nuonco/nuon-runner-go/models"

	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

type handlerState struct {
	// set during the fetch/validate phase
	plan                              *plantypes.DeployPlan
	appCfg                            *models.AppAppConfig
	kubernetesManifestComponentConfig *models.AppKubernetesManifestComponentConfig
	previousDeployResources           *string

	jobExecutionID string
	jobID          string
	timeout        time.Duration

	outputs map[string]interface{}

	// add validated manifest here
	kubeClient *kubernetesClient

	// OCI artifact archive (for pulling manifest from registry)
	arch   ociarchive.Archive
	srcCfg *configs.OCIRegistryRepository
	srcTag string
}

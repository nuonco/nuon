package kubernetes_manifest

import (
	"time"

	"github.com/nuonco/nuon-runner-go/models"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
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
}

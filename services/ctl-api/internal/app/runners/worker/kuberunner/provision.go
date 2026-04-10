package runner

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/workflows/types/executors"
)

const (
	defaultHelmOperationTimeout time.Duration = time.Minute * 10
)

type ProvisionRunnerRequestImage struct {
	URL string `validate:"required"`
	Tag string `validate:"required"`
}

type ProvisionRunnerRequest struct {
	RunnerID                 string                      `validate:"required"`
	APIURL                   string                      `validate:"required"`
	APIToken                 string                      `validate:"required"`
	Image                    ProvisionRunnerRequestImage `validate:"required"`
	RunnerServiceAccountName string                      `validate:"required"`

	// CloudProvider determines how the runner authenticates to cloud resources.
	// "gcp" uses GKE Workload Identity, "azure" uses AKS Workload Identity, anything else uses AWS IRSA.
	CloudProvider string

	// RunnerIAMRole is the AWS IAM role ARN (for AWS/IRSA), GCP service account email (for GCP/Workload Identity), or Azure managed identity client ID (for Azure/Workload Identity).
	RunnerIAMRole string
}

type ProvisionRunnerResponse struct{}

// @temporal-gen-v2 workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{ .CallerID }}-provision-runner
func (w Wkflow) ProvisionRunner(ctx workflow.Context, req *ProvisionRunnerRequest) (*executors.ProvisionRunnerResponse, error) {
	clusterInfo := w.getClusterInfo()

	if _, err := AwaitInstallOrUpgrade(ctx, &InstallOrUpgradeRequest{
		ClusterInfo: clusterInfo,
		Image:       req.Image,

		Namespace:                req.RunnerID,
		Timeout:                  defaultHelmOperationTimeout,
		RunnerID:                 req.RunnerID,
		RunnerServiceAccountName: req.RunnerServiceAccountName,
		RunnerIAMRole:            req.RunnerIAMRole,
		CloudProvider:            req.CloudProvider,
		APIToken:                 req.APIToken,
		APIURL:                   req.APIURL,
	}); err != nil {
		return nil, fmt.Errorf("unable to provision runner: %w", err)
	}

	return &executors.ProvisionRunnerResponse{}, nil
}

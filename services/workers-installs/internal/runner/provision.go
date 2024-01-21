package runner

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(v *validator.Validate, cfg workers.Config) wkflow {
	return wkflow{
		v:   v,
		cfg: cfg,
		act: NewActivities(nil, workers.Config{}),
		clusterInfo: kube.ClusterInfo{
			ID:             cfg.OrgsK8sClusterID,
			Endpoint:       cfg.OrgsK8sPublicEndpoint,
			CAData:         cfg.OrgsK8sCAData,
			TrustedRoleARN: cfg.OrgsK8sRoleArn,
		},
	}
}

type wkflow struct {
	v           *validator.Validate
	cfg         workers.Config
	act         *Activities
	clusterInfo kube.ClusterInfo
}

// Provision is a workflow that creates an app install sandbox using terraform
//
//nolint:funlen
func (w wkflow) ProvisionRunner(ctx workflow.Context, req *runnerv1.ProvisionRunnerRequest) (*runnerv1.ProvisionRunnerResponse, error) {
	resp := &runnerv1.ProvisionRunnerResponse{}
	l := workflow.GetLogger(ctx)

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	orgServerAddr := client.DefaultOrgServerAddress(w.cfg.OrgServerRootDomain, req.OrgId)
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// create waypoint project
	cwpReq := CreateWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          w.clusterInfo,
	}
	_, err := w.createWaypointProject(ctx, cwpReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint project: %w", err)
		return resp, err
	}

	// create waypoint workspace
	cwwReq := CreateWaypointWorkspaceRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          w.clusterInfo,
	}
	_, err = w.createWaypointWorkspace(ctx, cwwReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint workspace: %w", err)
		return resp, err
	}

	switch req.RunnerType {
	case installsv1.RunnerType_RUNNER_TYPE_AWS_ECS:
		if err := w.installECSRunner(ctx, req); err != nil {
			return resp, fmt.Errorf("unable to install ecs runner: %w", err)
		}
	case installsv1.RunnerType_RUNNER_TYPE_AWS_EKS:
		if err := w.installEKSRunner(ctx, req); err != nil {
			return resp, fmt.Errorf("unable to install eks runner: %w", err)
		}
	default:
		return resp, fmt.Errorf("unsupported runner type")
	}

	l.Debug("finished provisioning", "response", resp)
	return resp, nil
}

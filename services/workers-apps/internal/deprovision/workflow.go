package deprovision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	appv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
	workers "github.com/powertoolsdev/mono/services/workers-apps/internal"
)

type Workflow struct {
	cfg         workers.Config
	v           *validator.Validate
	act         *Activities
	clusterInfo kube.ClusterInfo
}

func NewWorkflow(v *validator.Validate, cfg workers.Config) Workflow {
	return Workflow{
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

func (w Workflow) Deprovision(ctx workflow.Context, req *appv1.DeprovisionRequest) (*appv1.DeprovisionResponse, error) {
	resp := appv1.DeprovisionResponse{}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	orgServerAddr := client.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgId)
	dwpReq := DestroyWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.WaypointTokenNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		AppID:                req.AppId,
		ClusterInfo:          w.clusterInfo,
	}

	// NOTE(jm): in some cases, it is possible for an org to not be active/alive, which is why the app is being torn
	// down. In that case, we only try this 5 times, and trust that the event loops will handle the logic.
	ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		MaximumAttempts: 5,
	})
	var dwpResp DestroyWaypointProjectResponse
	err := w.execWaypointActivity(ctx, w.act.DestroyWaypointProject, dwpReq, &dwpResp)
	if err != nil {
		err = fmt.Errorf("failed to destroy waypoint project: %w", err)
		return &resp, err
	}

	return &resp, nil
}

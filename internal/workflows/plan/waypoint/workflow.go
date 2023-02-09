package plan

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-waypoint"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	workers "github.com/powertoolsdev/workers-executors/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
)

func configureActivityOptions(ctx workflow.Context) workflow.Context {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	return workflow.WithActivityOptions(ctx, activityOpts)
}

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) CreatePlan(ctx workflow.Context, req *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities()

	if err := req.Validate(); err != nil {
		return resp, err
	}

	longIDs, err := shortid.ToUUIDs(req.OrgId, req.AppId, req.DeploymentId)
	if err != nil {
		return resp, fmt.Errorf("invalid shortids: %w", err)
	}
	var installID uuid.UUID
	if req.InstallId != "" {
		installID, err = shortid.ToUUID(req.InstallId)
		if err != nil {
			return resp, fmt.Errorf("invalid install shortid: %w", err)
		}
	}

	cpReq := &planactivitiesv1.CreatePlanRequest{
		Type: req.Type,
		Metadata: &planv1.Metadata{
			OrgId:             longIDs[0].String(),
			OrgShortId:        req.OrgId,
			AppShortId:        req.AppId,
			AppId:             longIDs[1].String(),
			DeploymentShortId: req.DeploymentId,
			DeploymentId:      longIDs[2].String(),
			InstallShortId:    req.InstallId,
			InstallId:         installID.String(),
		},
		OrgMetadata: &planv1.OrgMetadata{
			EcrRegion:      w.cfg.OrgsECRRegion,
			EcrRegistryId:  w.cfg.OrgsECRRegistryID,
			EcrRegistryArn: w.cfg.OrgsECRRegistryARN,
			Buckets: &planv1.OrgBuckets{
				DeploymentsBucket:   w.cfg.DeploymentsBucket,
				InstallationsBucket: w.cfg.InstallationsBucket,
				OrgsBucket:          w.cfg.OrgsBucket,
				InstancesBucket:     w.cfg.InstancesBucket,
			},
			WaypointServer: &planv1.WaypointServerRef{
				Address:              waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgId),
				TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
				TokenSecretName:      fmt.Sprintf(w.cfg.WaypointTokenSecretTemplate, req.OrgId),
			},
			IamRoleArns: &planv1.OrgIAMRoleArns{
				DeploymentsRoleArn:   fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, req.OrgId),
				InstallationsRoleArn: fmt.Sprintf(w.cfg.OrgsInstallationsRoleTemplate, req.OrgId),
				OdrRoleArn:           fmt.Sprintf(w.cfg.OrgsOdrRoleTemplate, req.OrgId),
				InstancesRoleArn:     fmt.Sprintf(w.cfg.OrgsInstancesRoleTemplate, req.OrgId),
				InstallerRoleArn:     fmt.Sprintf(w.cfg.OrgsInstallerRoleTemplate, req.OrgId),
				OrgsRoleArn:          fmt.Sprintf(w.cfg.OrgsOrgsRoleTemplate, req.OrgId),
			},
		},
		Component: req.Component,
	}
	cpResp, err := execCreatePlan(ctx, act, cpReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create plan: %w", err)
	}
	resp.Plan = cpResp.Plan

	l.Debug("successfully created plan for build")
	return resp, nil
}

func execCreatePlan(
	ctx workflow.Context,
	act *Activities,
	req *planactivitiesv1.CreatePlanRequest,
) (*planactivitiesv1.CreatePlanResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &planactivitiesv1.CreatePlanResponse{}

	l.Debug("executing create plan activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.CreatePlanAct, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

package plan

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-uploader"
	"github.com/powertoolsdev/go-waypoint"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners"
	waypointplanners "github.com/powertoolsdev/workers-executors/internal/planners/waypoint"
	waypointbuild "github.com/powertoolsdev/workers-executors/internal/planners/waypoint/build"
	waypointdeploy "github.com/powertoolsdev/workers-executors/internal/planners/waypoint/deploy"
	waypointsync "github.com/powertoolsdev/workers-executors/internal/planners/waypoint/sync"
)

func (w *wkflow) getComponentPlanRequest(typ planv1.PlanType, req *planv1.Component) (*planactivitiesv1.CreateComponentPlan, error) {
	longIDs, err := shortid.ToUUIDs(req.OrgId, req.AppId, req.DeploymentId)
	if err != nil {
		return nil, fmt.Errorf("invalid shortids: %w", err)
	}

	var installID uuid.UUID
	if req.InstallId != "" {
		installID, err = shortid.ToUUID(req.InstallId)
		if err != nil {
			return nil, fmt.Errorf("invalid install shortid: %w", err)
		}
	}

	return &planactivitiesv1.CreateComponentPlan{
		Type: typ,
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
	}, nil
}

func (a *Activities) getComponentPlanner(req *planactivitiesv1.CreateComponentPlan) (planners.Planner, error) {
	waypointOpts := []waypointplanners.PlannerOption{
		waypointplanners.WithComponent(req.Component),
		waypointplanners.WithMetadata(req.Metadata),
		waypointplanners.WithOrgMetadata(req.OrgMetadata),
	}

	switch req.Type {
	case planv1.PlanType_PLAN_TYPE_WAYPOINT_BUILD:
		return waypointbuild.New(a.v, waypointOpts...)
	case planv1.PlanType_PLAN_TYPE_WAYPOINT_SYNC_IMAGE:
		return waypointsync.New(a.v, waypointOpts...)
	case planv1.PlanType_PLAN_TYPE_WAYPOINT_DEPLOY:
		return waypointdeploy.New(a.v, waypointOpts...)
	default:
		return nil, fmt.Errorf("unsupported plan type: %s", req.Type)
	}
}

func (a *Activities) CreateComponentPlan(
	ctx context.Context,
	req *planactivitiesv1.CreateComponentPlan,
) (*planactivitiesv1.CreatePlanResponse, error) {
	resp := &planactivitiesv1.CreatePlanResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	planner, err := a.getComponentPlanner(req)
	if err != nil {
		return resp, fmt.Errorf("unable to get planner: %w", err)
	}

	plan, err := planner.Plan(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get plan: %w", err)
	}

	planRef := &planv1.PlanRef{
		Bucket:              req.OrgMetadata.Buckets.DeploymentsBucket,
		BucketAssumeRoleArn: req.OrgMetadata.IamRoleArns.DeploymentsRoleArn,
		BucketKey:           filepath.Join(planner.Prefix(), planKey),
	}

	// create upload client
	uploadClient, err := uploader.NewS3Uploader(a.v, uploader.WithBucketName(planRef.Bucket),
		uploader.WithAssumeSessionName("workers-executors"),
		uploader.WithAssumeRoleARN(planRef.BucketAssumeRoleArn))
	if err != nil {
		return resp, fmt.Errorf("unable to get uploader: %w", err)
	}

	err = a.planUploader.uploadPlan(ctx, uploadClient, planRef, plan)
	if err != nil {
		return resp, fmt.Errorf("unable to upload plan: %w", err)
	}
	resp.Plan = planRef

	return resp, nil
}

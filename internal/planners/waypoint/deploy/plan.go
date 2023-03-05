package deploy

import (
	"context"
	"fmt"
	"path/filepath"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint/configs"
)

const (
	defaultBuildTimeoutSeconds uint64 = 3600
)

func (p *planner) getBasePlan() *planv1.WaypointPlan {
	jobTyp := planv1.WaypointJobType_WAYPOINT_JOB_TYPE_DEPLOY_ARTIFACT
	if p.Component.BuildCfg.GetNoop() != nil {
		jobTyp = planv1.WaypointJobType_WAYPOINT_JOB_TYPE_DEPLOY
	}

	return &planv1.WaypointPlan{
		Metadata:       p.Metadata,
		WaypointServer: p.OrgMetadata.WaypointServer,
		EcrRepositoryRef: &planv1.ECRRepositoryRef{
			RepositoryName: p.Metadata.InstallShortId,
			Tag:            p.Metadata.DeploymentShortId,
		},
		WaypointRef: &planv1.WaypointRef{
			Project:              p.Metadata.InstallShortId,
			Workspace:            p.Metadata.InstallShortId,
			App:                  p.Component.Id,
			SingletonId:          fmt.Sprintf("%s-%s-%s", p.Metadata.DeploymentShortId, p.Metadata.InstallShortId, phaseName),
			Labels:               waypoint.DefaultLabels(p.Metadata, p.Component.Id, phaseName),
			RunnerId:             p.Metadata.InstallShortId,
			OnDemandRunnerConfig: p.Metadata.InstallShortId,
			JobTimeoutSeconds:    defaultBuildTimeoutSeconds,
			JobType:              jobTyp,
		},
		Outputs: &planv1.Outputs{
			Bucket:              p.OrgMetadata.Buckets.DeploymentsBucket,
			BucketPrefix:        p.Prefix(),
			BucketAssumeRoleArn: p.OrgMetadata.IamRoleArns.DeploymentsRoleArn,

			// TODO(jm): these aren't being used until we've fully implemented the executor
			LogsKey:     filepath.Join(p.Prefix(), "logs.txt"),
			EventsKey:   filepath.Join(p.Prefix(), "events.json"),
			ArtifactKey: filepath.Join(p.Prefix(), "artifacts.json"),
		},
		GitSource: &planv1.GitSource{
			Url: "https://github.com/jonmorehouse/empty",
		},
		Component: p.Component,
	}
}

func (p *planner) Plan(ctx context.Context) (*planv1.Plan, error) {
	plan := p.getBasePlan()

	cfg, err := configs.NewBasicDeploy(p.V,
		configs.WithEcrRef(plan.EcrRepositoryRef),
		configs.WithWaypointRef(plan.WaypointRef),
		configs.WithComponent(p.Component),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get plan: %w", err)
	}

	waypointCfg, waypointCfgFmt, err := cfg.Render()
	if err != nil {
		return nil, fmt.Errorf("unable to render config: %w", err)
	}
	plan.WaypointRef.HclConfig = string(waypointCfg)
	plan.WaypointRef.HclConfigFormat = waypointCfgFmt.String()

	return &planv1.Plan{Actual: &planv1.Plan_WaypointPlan{WaypointPlan: plan}}, nil
}

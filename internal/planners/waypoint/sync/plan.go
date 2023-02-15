package sync

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

func (p *planner) Plan(ctx context.Context) (*planv1.Plan, error) {
	plan := &planv1.WaypointPlan{
		Metadata:       p.Metadata,
		WaypointServer: p.OrgMetadata.WaypointServer,
		EcrRepositoryRef: &planv1.ECRRepositoryRef{
			RepositoryName: p.Metadata.InstallShortId,
			Tag:            p.Metadata.DeploymentShortId,
			// TODO(jm): we don't have a great way of knowing what region the customer install is using this
			// deep in this stage. Eventually, we would ideally fetch this information from `orgs-api`, but
			// for now just hard code us-west-2
			Region: "us-west-2",
		},
		WaypointRef: &planv1.WaypointRef{
			Project:              p.Metadata.InstallShortId,
			Workspace:            p.Metadata.InstallShortId,
			App:                  p.Component.Name,
			SingletonId:          fmt.Sprintf("%s-%s", p.Metadata.DeploymentShortId, p.Component.Name),
			Labels:               waypoint.DefaultLabels(p.Metadata, p.Component.Name, phaseName),
			RunnerId:             p.Metadata.InstallShortId,
			OnDemandRunnerConfig: p.Metadata.InstallShortId,
			JobTimeoutSeconds:    defaultBuildTimeoutSeconds,
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
		Component: p.Component,
	}

	srcImage, err := p.getSourceImage(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get source image: %w", err)
	}

	builder, err := configs.NewSyncImageBuilder(p.V,
		configs.WithEcrRef(plan.EcrRepositoryRef),
		configs.WithWaypointRef(plan.WaypointRef),
		configs.WithSourceImage(srcImage),
		configs.WithComponent(p.Component),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create builder: %w", err)
	}

	cfg, cfgFmt, err := builder.Render()
	if err != nil {
		return nil, fmt.Errorf("unable to render config: %w", err)
	}
	plan.WaypointRef.HclConfig = string(cfg)
	plan.WaypointRef.HclConfigFormat = cfgFmt.String()

	return &planv1.Plan{Actual: &planv1.Plan_WaypointPlan{WaypointPlan: plan}}, nil
}

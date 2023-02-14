package build

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
	ecrRepoName := fmt.Sprintf("%s/%s", p.Metadata.OrgShortId, p.Metadata.AppShortId)
	ecrRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", p.OrgMetadata.EcrRegistryId,
		p.OrgMetadata.EcrRegion, ecrRepoName)

	plan := &planv1.WaypointPlan{
		Metadata: p.Metadata,
		WaypointServer: &planv1.WaypointServerRef{
			Address:              p.OrgMetadata.WaypointServer.Address,
			TokenSecretNamespace: p.OrgMetadata.WaypointServer.TokenSecretNamespace,
			TokenSecretName:      p.OrgMetadata.WaypointServer.TokenSecretName,
		},
		EcrRepositoryRef: &planv1.ECRRepositoryRef{
			RegistryId:     p.OrgMetadata.EcrRegistryId,
			RepositoryName: ecrRepoName,
			RepositoryArn:  fmt.Sprintf("%s/%s", p.OrgMetadata.EcrRegistryArn, ecrRepoName),
			RepositoryUri:  ecrRepoURI,
			Tag:            p.Metadata.DeploymentShortId,
			Region:         p.OrgMetadata.EcrRegion,
		},
		WaypointRef: &planv1.WaypointRef{
			Project:              p.Metadata.AppShortId,
			Workspace:            p.Metadata.AppShortId,
			App:                  p.Component.Name,
			SingletonId:          fmt.Sprintf("%s-%s", p.Metadata.DeploymentShortId, p.Component.Name),
			Labels:               waypoint.DefaultLabels(p.Metadata, p.Component.Name, phaseName),
			RunnerId:             p.Metadata.OrgShortId,
			OnDemandRunnerConfig: p.Metadata.OrgShortId,
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

	// create builder which will render the waypoint config
	builder, err := configs.NewHttpbinBuildBuilder(p.V,
		configs.WithComponent(p.Component),
		configs.WithEcrRef(plan.EcrRepositoryRef),
		configs.WithWaypointRef(plan.WaypointRef),
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

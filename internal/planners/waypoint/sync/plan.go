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

func (p *planner) GetPlan(ctx context.Context) (*planv1.WaypointPlan, error) {
	plan := &planv1.WaypointPlan{
		Metadata: p.Metadata,
		// TODO(jm): we should probably just reuse the waypoint server ref for both of these, as they are
		// identical
		WaypointServer: &planv1.WaypointServerRef{
			Address:              p.OrgMetadata.WaypointServer.Address,
			TokenSecretNamespace: p.OrgMetadata.WaypointServer.TokenSecretNamespace,
			TokenSecretName:      p.OrgMetadata.WaypointServer.TokenSecretName,
		},
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
			BucketPrefix:        p.getPrefix(),
			BucketAssumeRoleArn: p.OrgMetadata.IamRoleArns.DeploymentsRoleArn,

			// TODO(jm): these aren't being used until we've fully implemented the executor
			LogsKey:     filepath.Join(p.getPrefix(), "logs.txt"),
			EventsKey:   filepath.Join(p.getPrefix(), "events.json"),
			ArtifactKey: filepath.Join(p.getPrefix(), "artifacts.json"),
		},
		Component: p.Component,
	}

	// TODO(jm):pass in correct values to builders
	//create builder which will render the waypoint config
	//ecrRepoName := fmt.Sprintf("%s/%s", p.Metadata.OrgShortId, p.Metadata.AppShortId)
	//ecrRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", p.OrgMetadata.EcrRegistryId,
	//p.OrgMetadata.EcrRegion, ecrRepoName)
	builder, err := configs.NewSyncImageBuilder(p.V,
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

	return plan, nil
}

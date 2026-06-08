package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/distribution/reference"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createContainerImageBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.ContainerImagePullPlan, error) {
	srcRepo, err := p.getSourceRepository(bld.ComponentConfigConnection.ExternalImageComponentConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get source repository")
	}

	plan := &plantypes.ContainerImagePullPlan{
		Image:        bld.ComponentConfigConnection.ExternalImageComponentConfig.ImageURL,
		Tag:          bld.ComponentConfigConnection.ExternalImageComponentConfig.Tag,
		UpdatePolicy: bld.ComponentConfigConnection.ExternalImageComponentConfig.UpdatePolicy,

		RepoCfg: srcRepo,
	}

	// Look up the most recent prior Active build's SourceDigest so the runner
	// can detect a no-op (upstream digest unchanged) and skip the artifact
	// push. Failure here is non-fatal — without the hint the runner just runs
	// a normal copy.
	prior, err := activities.AwaitGetPreviousActiveBuildSourceDigest(ctx, activities.GetPreviousActiveBuildSourceDigestRequest{
		ComponentID:    bld.ComponentConfigConnection.ComponentID,
		ExcludeBuildID: bld.ID,
	})
	if err != nil {
		if l, lerr := log.WorkflowLogger(ctx); lerr == nil {
			l.Warn("unable to look up previous source digest for build dedup",
				zap.String("component_id", bld.ComponentConfigConnection.ComponentID),
				zap.String("build_id", bld.ID),
				zap.Error(err))
		}
	} else if prior != nil {
		plan.PreviousSourceDigest = prior.SourceDigest
	}

	return plan, nil
}

func (b *Planner) normalizeRepository(repo string) (string, error) {
	ref, err := reference.ParseAnyReference(repo)
	if err != nil {
		return "", fmt.Errorf("invalid reference: %w", err)
	}

	named, err := reference.ParseDockerRef(ref.String())
	if err != nil {
		return "", fmt.Errorf("unable to parse docker ref: %w", err)
	}

	host := reference.Domain(named)
	if host == "docker.io" {
		// The normalized name parse above will turn short names like "foo/bar"
		// into "docker.io/foo/bar". We return "docker.io" and let oras-go
		// handle the mapping to "registry-1.docker.io" internally.
		// Using "index.docker.io" breaks the anonymous bearer token flow.
		return "docker.io", nil
	}

	// by default, if a reference is fully resolved, we just use the repository name
	return "", nil
}

func (b *Planner) getSourceRepository(cfg *app.ExternalImageComponentConfig) (*configs.OCIRegistryRepository, error) {
	loginServer, err := b.normalizeRepository(cfg.ImageURL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to normalize repository")
	}

	if cfg.AWSECRImageConfig != nil {
		return &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeECR,
			Repository:   cfg.ImageURL,
			Region:       cfg.AWSECRImageConfig.AWSRegion,

			ECRAuth: &credentials.Config{
				Region: cfg.AWSECRImageConfig.AWSRegion,
				AssumeRole: &credentials.AssumeRoleConfig{
					RoleARN:                cfg.AWSECRImageConfig.IAMRoleARN,
					SessionName:            "container-image-build",
					SessionDurationSeconds: 30 * 60,
					UseGCPOIDC:             b.cloudProvider == "gcp",
				},
			},
		}, nil
	}

	if cfg.GCPGARImageConfig != nil {
		garLoginServer := fmt.Sprintf("%s-docker.pkg.dev", cfg.GCPGARImageConfig.GCPRegion)
		return &configs.OCIRegistryRepository{
			RegistryType:             configs.OCIRegistryTypeGAR,
			Repository:               cfg.ImageURL,
			Region:                   cfg.GCPGARImageConfig.GCPRegion,
			LoginServer:              garLoginServer,
			ServiceAccountEmail:      cfg.GCPGARImageConfig.ServiceAccountEmail,
			WorkloadIdentityProvider: cfg.GCPGARImageConfig.WorkloadIdentityProvider,
		}, nil
	}

	if cfg.AzureACRImageConfig != nil {
		return &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeACR,
			Repository:   cfg.ImageURL,
			LoginServer:  cfg.AzureACRImageConfig.RegistryURL,
			ACRAuth: &azurecredentials.Config{
				UseDefault: true,
			},
		}, nil
	}

	return &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,

		Repository:  cfg.ImageURL,
		LoginServer: loginServer,

		OCIAuth: &configs.OCIRegistryAuth{},
	}, nil
}

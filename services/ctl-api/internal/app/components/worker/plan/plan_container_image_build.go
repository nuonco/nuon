package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/distribution/reference"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (p *Planner) createContainerImageBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.ContainerImagePullPlan, error) {
	srcRepo, err := p.getSourceRepository(bld.ComponentConfigConnection.ExternalImageComponentConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get source repository")
	}

	return &plantypes.ContainerImagePullPlan{
		Image: bld.ComponentConfigConnection.ExternalImageComponentConfig.ImageURL,
		Tag:   bld.ComponentConfigConnection.ExternalImageComponentConfig.Tag,

		RepoCfg: srcRepo,
	}, nil
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
		// into "docker.io/foo/bar" but the actual registry host for these
		// is "index.docker.io".
		return "index.docker.io", nil
	}

	// by default, if a reference is fully resolved, we just use the repository name
	return "", nil
}

func (b *Planner) getSourceRepository(cfg *app.ExternalImageComponentConfig) (*configs.OCIRegistryRepository, error) {
	loginServer, err := b.normalizeRepository(cfg.ImageURL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to normalize repository")
	}

	// aws ecr repository
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
				},
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

package build

import (
	"context"
	"fmt"

	buildv1 "github.com/powertoolsdev/protos/components/generated/types/build/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	awsecrauthorization "github.com/powertoolsdev/workers-executors/internal/aws-ecr-authorization"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint/configs"
)

const (
	defaultAssumeRoleSessionName string = "workers-executors-planner"
)

// getExternalImagePlan returns a waypoint plan for doing a build of an external image
func (p *planner) getExternalImagePlan(ctx context.Context, cfg *buildv1.Config_ExternalImageCfg) (*planv1.WaypointPlan, error) {
	plan := p.getBasePlan()

	baseOpts := []configs.Option{
		configs.WithComponent(p.Component),
		configs.WithEcrRef(plan.EcrRepositoryRef),
		configs.WithWaypointRef(plan.WaypointRef),
	}

	var (
		builder configs.Builder
		err     error
	)
	switch authCfg := cfg.ExternalImageCfg.AuthCfg.Cfg.(type) {
	case *buildv1.ExternalImageAuthConfig_AwsIamAuthCfg:
		//nolint:govet
		ecr, err := awsecrauthorization.New(p.V,
			awsecrauthorization.WithAssumeRoleArn(authCfg.AwsIamAuthCfg.IamRoleArn),
			awsecrauthorization.WithAssumeRoleSessionName(defaultAssumeRoleSessionName),
			awsecrauthorization.WithImageURL(cfg.ExternalImageCfg.OciImageUrl),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create ecrauthorizer for private docker pull build: %w", err)
		}

		ecrAuth, err := ecr.GetAuthorization(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
		}

		builder, err = configs.NewPrivateDockerPullBuild(p.V,
			append(baseOpts,
				configs.WithPrivateImageSource(&configs.PrivateImageSource{
					RegistryToken: ecrAuth.RegistryToken,
					ServerAddress: ecrAuth.ServerAddress,
					Username:      ecrAuth.Username,
					Image:         cfg.ExternalImageCfg.OciImageUrl,
					Tag:           cfg.ExternalImageCfg.Tag,
				}))...,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to build private docker pull build: %w", err)
		}
	case *buildv1.ExternalImageAuthConfig_PublicAuthCfg:
		builder, err = configs.NewPublicDockerPullBuild(p.V,
			append(baseOpts,
				configs.WithPublicImageSource(&configs.PublicImageSource{
					Image: cfg.ExternalImageCfg.OciImageUrl,
					Tag:   cfg.ExternalImageCfg.Tag,
				}))...,
		)
	default:
		return nil, fmt.Errorf("invalid external image config")
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create external image builder: %w", err)
	}

	waypointCfg, cfgFmt, err := builder.Render()
	if err != nil {
		return nil, fmt.Errorf("unable to render waypoint config: %w", err)
	}
	plan.WaypointRef.HclConfig = string(waypointCfg)
	plan.WaypointRef.HclConfigFormat = cfgFmt.String()
	return plan, nil
}

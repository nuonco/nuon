package protos

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) FromDeploy(deploy *app.InstallDeploy, installDeploys []app.InstallDeploy) (*componentv1.Component, error) {
	build := deploy.ComponentBuild

	compCfg := build.ComponentConfigConnection

	var (
		cfg *componentv1.Component
		err error
	)
	if compCfg.TerraformModuleComponentConfig != nil {
		cfg, err = c.ToTerraformModuleComponentConfig(
			compCfg.TerraformModuleComponentConfig,
			installDeploys,
			generics.FromPtrStr(build.GitRef),
			deploy,
		)
	}

	if compCfg.HelmComponentConfig != nil {
		cfg, err = c.ToHelmComponentConfig(
			compCfg.HelmComponentConfig,
			installDeploys,
			generics.FromPtrStr(build.GitRef))
	}

	if compCfg.ExternalImageComponentConfig != nil {
		cfg, err = c.ToExternalImageConfig(compCfg.ExternalImageComponentConfig, installDeploys)
	}

	if compCfg.JobComponentConfig != nil {
		cfg, err = c.ToJobConfig(compCfg.JobComponentConfig, installDeploys)
	}

	if compCfg.DockerBuildComponentConfig != nil {
		cfg, err = c.ToDockerBuildConfig(
			compCfg.DockerBuildComponentConfig,
			installDeploys,
			generics.FromPtrStr(build.GitRef))
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create component config: %w", err)
	}

	cfg.InstallInputs, err = c.toInstallInputs(deploy.InstallComponent.Install)
	if err != nil {
		return nil, fmt.Errorf("unable to create install inputs: %w", err)
	}

	return cfg, nil
}

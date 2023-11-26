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
	if compCfg.TerraformModuleComponentConfig != nil {
		return c.ToTerraformModuleComponentConfig(compCfg.TerraformModuleComponentConfig,
			installDeploys,
			generics.FromPtrStr(build.GitRef))
	}

	if compCfg.HelmComponentConfig != nil {
		return c.ToHelmComponentConfig(compCfg.HelmComponentConfig,
			installDeploys,
			generics.FromPtrStr(build.GitRef))
	}

	if compCfg.ExternalImageComponentConfig != nil {
		return c.ToExternalImageConfig(compCfg.ExternalImageComponentConfig, installDeploys)
	}

	if compCfg.JobComponentConfig != nil {
		return c.ToJobConfig(compCfg.JobComponentConfig, installDeploys)
	}

	if compCfg.DockerBuildComponentConfig != nil {
		return c.ToDockerBuildConfig(compCfg.DockerBuildComponentConfig,
			installDeploys,
			generics.FromPtrStr(build.GitRef))
	}

	return nil, fmt.Errorf("unable to convert component to proto component")
}

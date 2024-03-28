package protos

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) FromBuild(build *app.ComponentBuild) (*componentv1.Component, error) {
	compCfg := build.ComponentConfigConnection

	if compCfg.TerraformModuleComponentConfig != nil {
		return c.ToTerraformModuleComponentConfig(compCfg.TerraformModuleComponentConfig,
			nil,
			generics.FromPtrStr(build.GitRef))
	}

	if compCfg.HelmComponentConfig != nil {
		return c.ToHelmComponentConfig(compCfg.HelmComponentConfig,
			nil,
			generics.FromPtrStr(build.GitRef))
	}

	if compCfg.ExternalImageComponentConfig != nil {
		return c.ToExternalImageConfig(compCfg.ExternalImageComponentConfig, nil)
	}

	if compCfg.JobComponentConfig != nil {
		return c.ToJobConfig(compCfg.JobComponentConfig, nil)
	}

	if compCfg.DockerBuildComponentConfig != nil {
		return c.ToDockerBuildConfig(compCfg.DockerBuildComponentConfig,
			nil,
			generics.FromPtrStr(build.GitRef))
	}

	return nil, fmt.Errorf("unable to convert component to proto component")
}

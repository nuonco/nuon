package buildv1

import (
	vcsv1 "github.com/powertoolsdev/mono/pkg/types/components/vcs/v1"
)

func (c *Config) GetVcsCfg() *vcsv1.Config {
	switch cfg := c.Cfg.(type) {
	case *Config_DockerCfg:
		return cfg.DockerCfg.GetVcsCfg()
	case *Config_TerraformModuleCfg:
		return cfg.TerraformModuleCfg.GetVcsCfg()
	case *Config_HelmChartCfg:
		return cfg.HelmChartCfg.GetVcsCfg()
	case *Config_ExternalImageCfg:
	case *Config_Noop:
	default:
	}
	return nil
}

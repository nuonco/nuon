package config

type HelmValue struct {
	Name  string `toml:"name" mapstructure:"name"`
	Value string `toml:"value" mapstructure:"value"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type HelmChartComponentConfig struct {
	Name         string      `mapstructure:"name"`
	Dependencies []string    `mapstructure:"dependencies"`
	ChartName    string      `mapstructure:"chart_name"`
	Values       []HelmValue `mapstructure:"value"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo"`
}

func (t *HelmChartComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"
	return resource, nil
}

package config

type HelmValue struct {
	Name  string `toml:"name" mapstructure:"name"`
	Value string `toml:"value" mapstructure:"value"`
}

type HelmChartComponentConfig struct {
	ChartName string      `mapstructure:"chart_name" toml:"chart_name"`
	Values    []HelmValue `mapstructure:"values" toml:"values"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo" toml:"public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo" toml:"connected_repo"`
}

func (t *HelmChartComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	// blocks in terraform are singular, and repeated but in our config are plural.
	resource["value"] = resource["values"]
	delete(resource, "values")

	return resource, nil
}

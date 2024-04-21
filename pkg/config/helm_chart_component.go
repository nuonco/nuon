package config

type HelmValue struct {
	Name  string `toml:"name" mapstructure:"name,omitempty"`
	Value string `toml:"value" mapstructure:"value,omitempty"`
}

type HelmValuesFile struct {
	Contents string `toml:"contents" mapstructure:"contents,omitempty"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type HelmChartComponentConfig struct {
	Name         string            `mapstructure:"name,omitempty"`
	Dependencies []string          `mapstructure:"dependencies,omitempty"`
	ChartName    string            `mapstructure:"chart_name,omitempty"`
	Values       []HelmValue       `mapstructure:"value,omitempty"`
	ValuesMap    map[string]string `mapstructure:"values,omitempty"`

	ValuesFiles []HelmValuesFile `mapstructure:"values_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"`
}

func (t *HelmChartComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"
	delete(resource, "values")
	return resource, nil
}

func (t *HelmChartComponentConfig) parse() error {
	for k, v := range t.ValuesMap {
		t.Values = append(t.Values, HelmValue{
			Name:  k,
			Value: v,
		})
	}

	return nil
}

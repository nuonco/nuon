package variablesv1

func (c *Variables) EnvVars() *EnvVars {
	cfgs := make([]*EnvVar, 0)

	for _, cfg := range c.Variables {
		actual := cfg.GetEnvVar()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return &EnvVars{
		Env: cfgs,
	}
}

func (c *Variables) HelmValues() *HelmValues {
	cfgs := make([]*HelmValue, 0)

	for _, cfg := range c.Variables {
		actual := cfg.GetHelmValue()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return &HelmValues{
		Values: cfgs,
	}
}

func (c *Variables) HelmValueMaps() []*HelmValuesMap {
	cfgs := make([]*HelmValuesMap, 0)

	for _, cfg := range c.Variables {
		actual := cfg.GetHelmValuesMap()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return cfgs
}

func (c *Variables) WaypointVariables() *WaypointVariables {
	cfgs := make([]*WaypointVariable, 0)

	for _, cfg := range c.Variables {
		actual := cfg.GetWaypointVariable()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return &WaypointVariables{
		Variables: cfgs,
	}
}

func (c *Variables) TerraformVariables() *TerraformVariables {
	cfgs := make([]*TerraformVariable, 0)

	for _, cfg := range c.Variables {
		actual := cfg.GetTerraformVariable()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return &TerraformVariables{
		Variables: cfgs,
	}
}

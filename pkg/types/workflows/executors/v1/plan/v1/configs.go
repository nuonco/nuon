package planv1

func (c *Configs) EnvVars() *EnvVars {
	cfgs := make([]*EnvVar, 0)

	for _, cfg := range c.Configs {
		actual := cfg.GetEnvVar()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return &EnvVars{
		Env: cfgs,
	}
}

func (c *Configs) HelmValues() *HelmValues {
	cfgs := make([]*HelmValue, 0)

	for _, cfg := range c.Configs {
		actual := cfg.GetHelmValue()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return &HelmValues{
		Values: cfgs,
	}
}

func (c *Configs) HelmValueMaps() []*HelmValuesMap {
	cfgs := make([]*HelmValuesMap, 0)

	for _, cfg := range c.Configs {
		actual := cfg.GetHelmValuesMap()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return cfgs
}

func (c *Configs) WaypointVariables() *WaypointVariables {
	cfgs := make([]*WaypointVariable, 0)

	for _, cfg := range c.Configs {
		actual := cfg.GetWaypointVariable()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return &WaypointVariables{
		Variables: cfgs,
	}
}

func (c *Configs) TerraformVariables() *TerraformVariables {
	cfgs := make([]*TerraformVariable, 0)

	for _, cfg := range c.Configs {
		actual := cfg.GetTerraformVariable()
		if actual != nil {
			cfgs = append(cfgs, actual)
		}
	}

	return &TerraformVariables{
		Variables: cfgs,
	}
}

// ToMap converts a map of env vars into a mapping
func (e *EnvVars) ToMap() map[string]string {
	if e == nil {
		return nil
	}

	vals := make(map[string]string, 0)
	for _, envVar := range e.Env {
		vals[envVar.Name] = envVar.Value
	}

	return vals
}

// ToMap converts a map of waypoint variables into a mapping
func (w *WaypointVariables) ToMap() map[string]string {
	if w == nil {
		return nil
	}

	vals := make(map[string]string, 0)
	for _, wpVar := range w.Variables {
		vals[wpVar.Name] = wpVar.Value
	}

	return vals
}

// ToMap converts a map of terraform variables into a mapping
func (t *TerraformVariables) ToMap() map[string]string {
	if t == nil {
		return nil
	}

	vals := make(map[string]string, 0)
	for _, tfVar := range t.Variables {
		vals[tfVar.Name] = tfVar.Value
	}

	return vals
}

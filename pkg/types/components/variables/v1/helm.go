package variablesv1

func (e *HelmValue) SetValue(val string) {
	e.Value = val
}

func (e *HelmValues) ToMap() map[string]string {
	vals := make(map[string]string)
	for _, hv := range e.Values {
		vals[hv.Name] = hv.Value
	}

	return vals
}

// ToVariables converts a list env vars into a variables array
func (e *HelmValues) ToVariables() []*Variable {
	if e == nil {
		return nil
	}

	vars := make([]*Variable, 0)
	for _, envVar := range e.Values {
		ev := envVar
		vars = append(vars, &Variable{
			Actual: &Variable_HelmValue{
				HelmValue: ev,
			},
		})
	}

	return vars
}

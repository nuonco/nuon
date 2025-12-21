package variablesv1

func (e *EnvVar) SetValue(val string) {
	e.Value = val
}

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

// ToVariables converts a list env vars into a variables array
func (e *EnvVars) ToVariables() []*Variable {
	if e == nil {
		return nil
	}

	vars := make([]*Variable, 0)
	for _, envVar := range e.Env {
		ev := envVar
		vars = append(vars, &Variable{
			Actual: &Variable_EnvVar{
				EnvVar: ev,
			},
		})
	}
	return vars
}

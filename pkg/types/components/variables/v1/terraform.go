package variablesv1

func (t *TerraformVariable) SetValue(val string) {
	t.Value = val
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

func (e *TerraformVariables) ToVariables() []*Variable {
	if e == nil {
		return nil
	}

	vars := make([]*Variable, 0)
	for _, envVar := range e.Variables {
		ev := envVar
		vars = append(vars, &Variable{
			Actual: &Variable_TerraformVariable{
				TerraformVariable: ev,
			},
		})
	}

	return vars
}

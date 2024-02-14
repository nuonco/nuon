package variablesv1

// ToMap converts a map of terraform variables into a mapping
func (t *InstallInputs) ToMap() map[string]string {
	if t == nil {
		return nil
	}

	vals := make(map[string]string, 0)
	for _, input := range t.InstallInputs {
		vals[input.Name] = input.Value
	}

	return vals
}

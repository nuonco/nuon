package variablesv1

func (w *WaypointVariable) SetValue(val string) {
	w.Value = val
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

func (e *WaypointVariables) ToVariables() []*Variable {
	if e == nil {
		return nil
	}

	vars := make([]*Variable, 0)
	for _, envVar := range e.Variables {
		ev := envVar
		vars = append(vars, &Variable{
			Actual: &Variable_WaypointVariable{
				WaypointVariable: ev,
			},
		})
	}

	return vars
}

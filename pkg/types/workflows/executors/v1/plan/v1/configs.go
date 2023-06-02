package planv1

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

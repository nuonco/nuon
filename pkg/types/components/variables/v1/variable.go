package variablesv1

func (v *Variable) GetVariable() variable {
	switch act := v.Actual.(type) {
	case *Variable_EnvVar:
		return act.EnvVar
	case *Variable_HelmValue:
		return act.HelmValue
	case *Variable_TerraformVariable:
		return act.TerraformVariable
	case *Variable_WaypointVariable:
		return act.WaypointVariable
	default:
	}

	return nil
}

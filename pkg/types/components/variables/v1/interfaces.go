package variablesv1

// we enforce a couple interfaces to allow us to more easily work with different types of variables, in consistent ways
type variable interface {
	GetValue() string
	SetValue(string)
}

var _ variable = (*HelmValue)(nil)
var _ variable = (*EnvVar)(nil)
var _ variable = (*TerraformVariable)(nil)
var _ variable = (*WaypointVariable)(nil)

// variable lists are lists of the same variables
type variableList interface {
	ToVariables() []*Variable
	ToMap() map[string]string
}

var _ variableList = (*EnvVars)(nil)
var _ variableList = (*HelmValues)(nil)
var _ variableList = (*TerraformVariables)(nil)
var _ variableList = (*WaypointVariables)(nil)

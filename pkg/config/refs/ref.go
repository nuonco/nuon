package refs

type RefType string

const (
	RefTypeSandbox          RefType = "sandbox_outputs"
	RefTypeInstallStack     RefType = "install_stack_outputs"
	RefTypeComponents       RefType = "component_outputs"
	RefTypeComponentsNested RefType = "component_outputs_nested"
	RefTypeInputs           RefType = "inputs"
)

type Ref struct {
	Type RefType
	Name string
}

func (r Ref) String() string {
	return string(r.Type) + "." + r.Name
}

func (r Ref) Equal(other Ref) bool {
	return r.Type == other.Type && r.Name == other.Name
}

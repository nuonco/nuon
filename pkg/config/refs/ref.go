package refs

import (
	"fmt"
	"strings"
)

type RefType string

const (
	RefTypeSandbox       RefType = "sandbox"
	RefTypeInstallStack  RefType = "install_stack"
	RefTypeComponents    RefType = "component"
	RefTypeInputs        RefType = "inputs"
	RefTypeInstallInputs RefType = "install_inputs"
	RefTypeSecrets       RefType = "secrets"
	RefTypeActions       RefType = "actions"
)

type Ref struct {
	Type RefType

	Name  string
	Value string
	Input string
}

func (r Ref) String() string {
	return fmt.Sprintf("%s|%s|%s|%s", r.Type, r.Name, r.Value, r.Input)
}

func (r Ref) Equal(other Ref) bool {
	return r.Type == other.Type && r.Name == other.Name
}

func NewFromString(str string) Ref {
	pieces := strings.Split(str, "|")

	ref := Ref{}
	if len(pieces) >= 1 {
		ref.Type = RefType(pieces[0])
	}
	if len(pieces) >= 2 {
		ref.Name = pieces[1]
	}
	if len(pieces) >= 3 {
		ref.Value = pieces[2]
	}
	if len(pieces) >= 4 {
		ref.Input = pieces[3]
	}

	return ref
}

// Principal package contains operations related to nuon entity principals like components, actions, sandboxes etc
package principal

import (
	"fmt"
	"strings"
)

// Type represents the type of principal entity
type Type string

const (
	TypeComponent Type = "component"
	TypeSandbox   Type = "sandbox"
	TypeAction    Type = "action"
)

var ValidTypes = []Type{
	TypeComponent,
	TypeSandbox,
	TypeAction,
}

type Principal struct {
	Type Type
	Name string
}

func ParsePrincipal(principalStr string) (*Principal, error) {
	if !strings.HasPrefix(principalStr, "nuon::") {
		return nil, fmt.Errorf("principal must start with 'nuon::'")
	}

	remainder := strings.TrimPrefix(principalStr, "nuon::")

	// split by ":" to separate principal type and name
	parts := strings.SplitN(remainder, ":", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid principal format: %s", principalStr)
	}

	principalType := parts[0]

	var principalName string

	// check if there's a name part
	if len(parts) == 2 {
		principalName = parts[1]
	}

	if principalType == "" {
		return nil, fmt.Errorf("principalType cannot be empty, should be either component, action or sandbox")
	}

	return &Principal{
		Type: Type(principalType),
		Name: principalName,
	}, nil
}

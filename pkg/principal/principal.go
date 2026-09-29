package principal

import (
	"fmt"
	"strings"
)

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

	parts := strings.SplitN(remainder, ":", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid principal format: %s", principalStr)
	}

	principalType := parts[0]

	var principalName string

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

package build

import (
	"fmt"
	"slices"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type OperationRoleRuleInput struct {
	Principal string
	Operation app.OperationType
	Role      string
}

func OperationRoleRuleInputsFromConfig(roles *config.OperationRolesConfig) []OperationRoleRuleInput {
	if roles == nil {
		return nil
	}
	out := make([]OperationRoleRuleInput, 0, len(roles.RuleMatrix))
	for _, rule := range roles.RuleMatrix {
		out = append(out, OperationRoleRuleInput{
			Principal: rule.Principal,
			Operation: app.OperationType(rule.Operation),
			Role:      rule.RoleName,
		})
	}
	return out
}

func OperationRoleConfig(appID, appConfigID string) *app.AppOperationRoleConfig {
	return &app.AppOperationRoleConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
	}
}

// OperationRoleRules validates and builds the rules for an already-persisted
// operation role config.
func OperationRoleRules(rules []OperationRoleRuleInput, configID string) ([]*app.AppOperationRoleRule, error) {
	out := make([]*app.AppOperationRoleRule, 0, len(rules))
	for _, rule := range rules {
		if !slices.Contains(app.ValidOperations, rule.Operation) {
			return nil, fmt.Errorf("invalid operation type: %s", rule.Operation)
		}

		p, err := principal.ParsePrincipal(rule.Principal)
		if err != nil {
			return nil, fmt.Errorf("invalid principal %q: %w", rule.Principal, err)
		}

		out = append(out, &app.AppOperationRoleRule{
			AppOperationRoleConfigID: configID,
			PrincipalType:            p.Type,
			PrincipalName:            p.Name,
			Operation:                rule.Operation,
			Role:                     rule.Role,
		})
	}
	return out, nil
}

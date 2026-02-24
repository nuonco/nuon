package sync

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/generics"
)

func (s sync) getOperationRulesRequest() *models.ServiceCreateAppOperationRoleConfigRequest {
	rules := make([]*models.ServiceOperationRoleRuleRequest, 0, len(s.cfg.OperationRoles.RuleMatrix))
	for _, rule := range s.cfg.OperationRoles.RuleMatrix {
		rules = append(rules, &models.ServiceOperationRoleRuleRequest{
			Principal: generics.ToPtr(rule.Principal),
			Operation: generics.ToPtr(models.AppOperationType(rule.Operation)),
			Role:      generics.ToPtr(rule.RoleName),
		})
	}

	return &models.ServiceCreateAppOperationRoleConfigRequest{
		AppConfigID: generics.ToPtr(s.appConfigID),
		Rules:       rules,
	}
}

func (s sync) syncAppOperationRules(ctx context.Context, resource string) error {
	req := s.getOperationRulesRequest()
	_, err := s.apiClient.CreateAppOperationRoleConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}

package service

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *AppConfigTypesTestSuite) TestCreateConfigTypesRejectAppConfigFromAnotherApp() {
	ctx, _ := s.service.Seeder.EnsureAccount(context.Background(), s.T())
	otherOrgCtx, _ := s.service.Seeder.EnsureOrg(ctx, s.T())
	otherOrgApp := s.service.Seeder.CreateApp(otherOrgCtx, s.T())
	otherOrgConfig := s.service.Seeder.CreateBareAppConfig(otherOrgCtx, s.T(), otherOrgApp.ID)

	sameOrgCtx := cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	sameOrgApp := s.service.Seeder.CreateApp(sameOrgCtx, s.T())
	sameOrgConfig := s.service.Seeder.CreateBareAppConfig(sameOrgCtx, s.T(), sameOrgApp.ID)

	role := map[string]any{
		"name":         "role",
		"display_name": "role",
		"description":  "role",
		"policies":     []map[string]any{{"managed_policy_name": "AdministratorAccess"}},
	}
	endpoints := map[string]func(appConfigID string) map[string]any{
		"stack-configs": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "type": "aws-cloudformation", "name": "x", "description": "x"}
		},
		"runner-configs": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "type": "aws"}
		},
		"secrets-configs": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "secrets": []map[string]any{{"name": "s", "display_name": "s", "description": "s"}}}
		},
		"policies-configs": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "policies": []any{}}
		},
		"break-glass-configs": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "roles": []any{}}
		},
		"permissions-configs": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "provision_role": role, "deprovision_role": role, "maintenance_role": role}
		},
		"sandbox-configs": func(id string) map[string]any {
			return map[string]any{
				"app_config_id":         id,
				"variables":             map[string]any{},
				"env_vars":              map[string]any{},
				"terraform_version":     "1.7.5",
				"public_git_vcs_config": map[string]any{"repo": "https://github.com/nuonco/sandboxes", "directory": "aws-eks", "branch": "main"},
			}
		},
		"input-config": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "groups": map[string]any{}, "inputs": map[string]any{}}
		},
		"kubernetes-contexts-configs": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "contexts": []any{}}
		},
		"operation-role-configs": func(id string) map[string]any {
			return map[string]any{"app_config_id": id, "rules": []any{}}
		},
	}

	foreignConfigs := map[string]string{
		"app config in another org":            otherOrgConfig.ID,
		"app config of another app in the org": sameOrgConfig.ID,
	}

	for endpoint, body := range endpoints {
		for name, appConfigID := range foreignConfigs {
			s.Run(endpoint+" rejects "+name, func() {
				rr := s.makeRequest(http.MethodPost, "/v1/apps/"+s.testApp.ID+"/"+endpoint, body(appConfigID))
				if rr.Code != http.StatusNotFound {
					s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
				}
				require.Equal(s.T(), http.StatusNotFound, rr.Code)
			})
		}
	}
}

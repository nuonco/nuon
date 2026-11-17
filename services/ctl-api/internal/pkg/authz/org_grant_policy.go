package authz

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func orgGrantPolicy(orgID string, roleType app.RoleType, api permissions.API) ([]byte, error) {
	refs := []permissions.PermissionID{permissions.OrgRead, permissions.SlackConnect}
	switch roleType {
	case app.RoleTypeOrgReadOnly:
	case app.RoleTypeOrgAdmin:
		refs = []permissions.PermissionID{permissions.OrgRead, permissions.OrgWrite, permissions.ServiceAccountsRead, permissions.TokensRead, permissions.TrustPoliciesRead}
	case app.RoleTypeOrgSupport, app.RoleTypeInstaller:
		refs = []permissions.PermissionID{
			permissions.OrgRead, permissions.OrgUpdate, permissions.OrgDelete,
			permissions.OrgMemberAdd, permissions.OrgMemberRemove, permissions.OrgInviteCreate,
			permissions.OrgFeaturesUpdate, permissions.OrgWebhookCreate, permissions.OrgWebhookUpdate, permissions.OrgWebhookDelete,
			permissions.AppCreate, permissions.AppDelete, permissions.AppAutomationManage, permissions.AppDefinitionWrite,
			permissions.AppBundleCreate, permissions.AppBundleDownload, permissions.InstallCreate,
			permissions.InstallUpdate, permissions.InstallDelete, permissions.InstallProvision, permissions.InstallDeploy,
			permissions.InstallTeardown, permissions.InstallInputsUpdate, permissions.InstallConfigUpdate, permissions.InstallSecretsSync,
			permissions.InstallDriftResolve, permissions.InstallRunnerBootstrap, permissions.InstallCollectorManage,
			permissions.ActionExecute, permissions.ActionAdhocExecute, permissions.RunbookExecute,
			permissions.NotebookCreate, permissions.NotebookUpdate, permissions.NotebookDelete, permissions.NotebookExecute,
			permissions.WorkflowCancel, permissions.WorkflowRetry, permissions.WorkflowApprove, permissions.RunnerJobCancel,
			permissions.RunnerControl, permissions.RunnerSettingsUpdate, permissions.RunnerTokensPrune,
			permissions.TerraformCreate, permissions.TerraformDelete, permissions.TerraformLock, permissions.TerraformStateWrite,
			permissions.ActionCreate, permissions.ActionUpdate, permissions.ActionDelete, permissions.ActionConfigCreate,
			permissions.RunbookCreate, permissions.RunbookUpdate, permissions.RunbookDelete, permissions.RunbookConfigCreate,
			permissions.OnboardingManage, permissions.CloudConnectionCreate, permissions.CloudConnectionDelete, permissions.CloudConnectionVerify,
			permissions.ImageRepositoryCreate, permissions.ImageRepositoryDelete, permissions.VCSConnectionCreate,
			permissions.VCSConnectionDelete, permissions.VCSWebhookSubscribe, permissions.SlackConnect, permissions.SlackDisconnect,
			permissions.SlackSubscriptionWrite, permissions.TriggerCreate, permissions.TriggerDelete, permissions.TriggerCredentials,
			permissions.TriggerUpdate, permissions.TriggerReplay, permissions.TriggerRetry, permissions.ActionRunReport,
			permissions.StackReport, permissions.StackVersionReport, permissions.TelemetryTokenCreate,
			permissions.RunnerHealthReport, permissions.RunnerComponentHealthReport, permissions.RunnerProcessReport,
			permissions.RunnerTelemetryWrite, permissions.RunnerTelemetryTokenCreate, permissions.RunnerJobClaim,
			permissions.RunnerJobReport, permissions.TerraformStateDelete, permissions.HelmReleaseWrite,
			permissions.HelmReleaseDelete, permissions.LogStreamAppend,
		}
	default:
		return nil, fmt.Errorf("role %s has no customer grant recipe", roleType)
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return nil, err
	}
	document, err := catalog.NormalizePolicy(permissions.Policy{
		Format: permissions.PolicyFormatV4, API: api,
		Grants: []permissions.Grant{{Scope: permissions.GrantScope{Kind: permissions.KindOrg, IDs: []string{orgID}}, Permissions: refs}},
	})
	if err != nil {
		return nil, err
	}
	if _, err := catalog.Compile(document, orgID); err != nil {
		return nil, err
	}
	return json.Marshal(document)
}

func useOrgGrantPolicies(roles []app.Role) error {
	for i := range roles {
		role := &roles[i]
		if role.RoleType == app.RoleTypeRunner {
			policy, err := workloadGrantPolicy(role)
			if err != nil {
				return err
			}
			policy.API = ""
			document, err := json.Marshal(policy)
			if err != nil {
				return err
			}
			role.Policies[0].Document, role.Policies[0].Permissions = document, nil
			continue
		}
		if role.RoleType != app.RoleTypeOrgAdmin && role.RoleType != app.RoleTypeOrgReadOnly &&
			role.RoleType != app.RoleTypeOrgSupport && role.RoleType != app.RoleTypeInstaller {
			continue
		}
		document, err := orgGrantPolicy(role.OrgID.ValueString(), role.RoleType, "")
		if err != nil {
			return err
		}
		role.Policies[0].Document = document
		role.Policies[0].Permissions = nil
	}
	return nil
}

func isStandardOrgGrant(role app.Role) bool {
	if !role.Managed || len(role.Policies) != 1 || !role.Policies[0].HasGrantDocument() {
		return false
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return false
	}
	document, err := role.Policies[0].DecodeGrants(catalog)
	if err != nil || (document.API != "" && document.API != permissions.APICtl) {
		return false
	}
	expected, err := orgGrantPolicy(role.OrgID.ValueString(), role.RoleType, document.API)
	if err != nil {
		return false
	}
	actual, err := json.Marshal(document)
	return err == nil && bytes.Equal(expected, actual)
}

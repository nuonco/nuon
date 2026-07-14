package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	awsstacks "github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/aws"
	gcpstacks "github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/gcp"
)

// defaultGCPRunnerInitScript is the runner bootstrap script the GCP module
// fetches when the app's RunnerConfig doesn't pin one.
const defaultGCPRunnerInitScript = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/gcp/init.sh"

// buildInstallerSDKConfig renders the full install-stack configuration for an
// install: runner details, operation-role permissions/policies, break-glass and
// custom roles, install-input names, and secrets. It is the shared source of
// truth for the read-only config endpoint the Terraform provider's nuon_stack
// data source consumes.
func (s *service) buildInstallerSDKConfig(ctx context.Context, installID string) (*app.InstallerSDKConfig, error) {
	var install app.Install
	if res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("RunnerGroup.Runners").
		Preload("RunnerGroup.Settings").
		Where("id = ?", installID).
		First(&install); res.Error != nil {
		return nil, fmt.Errorf("load install: %w", res.Error)
	}

	// Per-runner-group setting wins; global config is only a safety net for
	// older installs that pre-date the per-group field being populated.
	runnerAPIURL := install.RunnerGroup.Settings.RunnerAPIURL
	if runnerAPIURL == "" {
		runnerAPIURL = s.cfg.RunnerAPIURL
	}
	if install.RunnerID == "" {
		return nil, fmt.Errorf("install %s has no runner — cannot build SDK config", installID)
	}
	if runnerAPIURL == "" {
		return nil, fmt.Errorf("install %s: runner_api_url empty (set RunnerGroupSettings.RunnerAPIURL for the install's runner group)", installID)
	}

	// GetFullAppConfig preloads PermissionsConfig, BreakGlassConfig,
	// InputConfig, SecretsConfig — same data the TF renderer walks.
	appCfg, err := s.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, true)
	if err != nil {
		return nil, fmt.Errorf("load full app config: %w", err)
	}
	if appCfg == nil {
		return nil, fmt.Errorf("install %s: app config not found", installID)
	}

	// Render `{{ .nuon.install.id }}` and friends in role names + policy
	// contents before passing to the SDK. Without this, IAM rejects policy docs
	// containing literal template syntax with "policy failed legacy parsing",
	// and role names get created with literal `{{` characters.
	installState, err := s.helpers.GetInstallState(ctx, installID, false, false)
	if err != nil {
		return nil, fmt.Errorf("get install state for template render: %w", err)
	}
	stateData, err := installState.AsMap()
	if err != nil {
		return nil, fmt.Errorf("install state as map: %w", err)
	}
	if err := render.RenderStruct(&appCfg.PermissionsConfig, stateData); err != nil {
		return nil, fmt.Errorf("render permissions config: %w", err)
	}
	if err := render.RenderStruct(&appCfg.BreakGlassConfig, stateData); err != nil {
		return nil, fmt.Errorf("render break-glass config: %w", err)
	}
	if err := render.RenderStruct(&appCfg.SecretsConfig, stateData); err != nil {
		return nil, fmt.Errorf("render secrets config: %w", err)
	}

	// Customer install inputs — names only. Values come from the per-install
	// inputs table at apply time, mirroring the TF tfvars contract which also
	// writes `"name" = ""` and lets the runner read values at runtime.
	var installInputs map[string]string
	var requiredInputs []string
	for _, in := range appCfg.InputConfig.AppInputs {
		if in.Source != app.AppInputSourceCustomer {
			continue
		}
		if installInputs == nil {
			installInputs = map[string]string{}
		}
		installInputs[in.Name] = ""
		if in.Required {
			requiredInputs = append(requiredInputs, in.Name)
		}
	}

	var autoGen []string
	var secrets map[string]app.InstallerSDKSecret
	for _, sec := range appCfg.SecretsConfig.Secrets {
		if sec.AutoGenerate {
			autoGen = append(autoGen, sec.Name)
			continue
		}
		if secrets == nil {
			secrets = map[string]app.InstallerSDKSecret{}
		}
		secrets[sec.Name] = app.InstallerSDKSecret{
			Description: sec.Description,
			Required:    sec.Required,
			Value:       sec.Default,
		}
	}

	cfg := &app.InstallerSDKConfig{
		InstallID:           install.ID,
		OrgID:               install.OrgID,
		AppID:               install.AppID,
		RunnerID:            install.RunnerID,
		RunnerAPIURL:        runnerAPIURL,
		InstallInputs:       installInputs,
		RequiredInputs:      requiredInputs,
		AutoGenerateSecrets: autoGen,
		Secrets:             secrets,
	}

	switch appCfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		if install.AWSAccount == nil || install.AWSAccount.Region == "" {
			return nil, fmt.Errorf("install %s has no AWS region; aws SDK provisioner requires it", installID)
		}

		provMPAs, maintMPAs, deprovMPAs := awsstacks.ExtractAWSStandardPermissionsRaw(appCfg)
		provDoc, maintDoc, deprovDoc, err := awsstacks.ExtractAWSStandardInlinePoliciesRaw(appCfg)
		if err != nil {
			return nil, fmt.Errorf("extract aws inline policies: %w", err)
		}
		breakGlass, err := awsstacks.ExtractAWSRolesFromListRaw(appCfg.BreakGlassConfig.Roles)
		if err != nil {
			return nil, fmt.Errorf("extract break-glass roles: %w", err)
		}
		customRoles, err := awsstacks.ExtractAWSRolesFromListRaw(appCfg.PermissionsConfig.CustomRoles)
		if err != nil {
			return nil, fmt.Errorf("extract custom roles: %w", err)
		}

		// Trust principal for operation roles. Empty config = empty list = SDK
		// falls back to account root, same as the TF module's
		// `control_plane_assume` default.
		var supportARNs []string
		if s.cfg.RunnerDefaultSupportIAMRole != "" {
			supportARNs = []string{s.cfg.RunnerDefaultSupportIAMRole}
		}

		// Cluster name: the install input "cluster_name" if set, else the
		// install ID. This becomes the kubernetes.io/cluster/<name> subnet tag
		// the EKS sandbox terraform expects.
		clusterName := install.ID
		if install.CurrentInstallInputs != nil {
			if v, ok := install.CurrentInstallInputs.Values["cluster_name"]; ok && v != nil && *v != "" {
				clusterName = *v
			}
		}

		cfg.Cloud = "aws"
		cfg.AWS = &app.InstallerSDKAWSConfig{
			Region:      install.AWSAccount.Region,
			ClusterName: clusterName,

			NuonSupportIAMRoleARNs: supportARNs,

			ProvisionManagedPolicyARNs:      provMPAs,
			ProvisionInlinePolicyDocument:   provDoc,
			MaintenanceManagedPolicyARNs:    maintMPAs,
			MaintenanceInlinePolicyDocument: maintDoc,
			DeprovisionManagedPolicyARNs:    deprovMPAs,
			DeprovisionInlinePolicyDocument: deprovDoc,

			// break-glass: enabled=false (created but disabled, matching TF's
			// `count` gate). custom: enabled=true.
			BreakGlassRoles: rolesToSDKConfigMap(breakGlass, false),
			CustomRoles:     rolesToSDKConfigMap(customRoles, true),
		}

	case app.AppRunnerTypeGCP:
		// NOTE: project + region are NOT known server-side — the customer
		// supplies them at provision time via the CLI. ctl-api only provides the
		// Nuon-generated inputs.

		// GCP provisions via the Terraform module, which authenticates the
		// runner with a real API token (no IID-based auth like AWS).
		token, err := s.runnersHelpers.CreateToken(ctx, install.RunnerID)
		if err != nil {
			return nil, fmt.Errorf("create runner token: %w", err)
		}
		initScriptURL := defaultGCPRunnerInitScript
		if appCfg.RunnerConfig.InitScriptURL != "" {
			initScriptURL = appCfg.RunnerConfig.InitScriptURL
		}

		prov, maint, deprov := gcpstacks.ExtractGCPStandardRolesRaw(appCfg)
		breakGlass := gcpstacks.ExtractGCPRolesRaw(appCfg.BreakGlassConfig.Roles)
		customRoles := gcpstacks.ExtractGCPRolesRaw(appCfg.PermissionsConfig.CustomRoles)

		cfg.Cloud = "gcp"
		cfg.GCP = &app.InstallerSDKGCPConfig{
			RunnerInitScriptURL: initScriptURL,
			RunnerAPIToken:      token.Token,

			ProvisionPermissions:      prov.Permissions,
			ProvisionPredefinedRole:   prov.PredefinedRole,
			MaintenancePermissions:    maint.Permissions,
			MaintenancePredefinedRole: maint.PredefinedRole,
			DeprovisionPermissions:    deprov.Permissions,
			DeprovisionPredefinedRole: deprov.PredefinedRole,

			ProvisionPolicies:   prov.Policies,
			MaintenancePolicies: maint.Policies,
			DeprovisionPolicies: deprov.Policies,

			BreakGlassRoles: gcpRolesToSDKMap(breakGlass, false),
			CustomRoles:     gcpRolesToSDKMap(customRoles, true),
		}

	default:
		return nil, fmt.Errorf("install %s: runner type %q is not supported by the SDK provisioner", installID, appCfg.RunnerConfig.Type)
	}

	return cfg, nil
}

func rolesToSDKConfigMap(rs []awsstacks.AWSRoleRaw, enabled bool) map[string]app.InstallerSDKRoleConfig {
	if len(rs) == 0 {
		return nil
	}
	out := make(map[string]app.InstallerSDKRoleConfig, len(rs))
	for _, r := range rs {
		out[r.Name] = app.InstallerSDKRoleConfig{
			InlinePolicyDocument: r.InlinePolicyDocument,
			ManagedPolicyARNs:    r.ManagedPolicyARNs,
			Enabled:              enabled,
		}
	}
	return out
}

func gcpRolesToSDKMap(rs []gcpstacks.GCPRoleRaw, enabled bool) map[string]app.InstallerSDKGCPRole {
	if len(rs) == 0 {
		return nil
	}
	out := make(map[string]app.InstallerSDKGCPRole, len(rs))
	for _, r := range rs {
		out[r.Name] = app.InstallerSDKGCPRole{
			Permissions:    r.Permissions,
			PredefinedRole: r.PredefinedRole,
			Enabled:        enabled,
			Policies:       r.Policies,
		}
	}
	return out
}

package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	awsstacks "github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/aws"
	gcpstacks "github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/gcp"
)

// defaultGCPRunnerInitScript is the runner bootstrap script the GCP module
// fetches when the app's RunnerConfig doesn't pin one.
const defaultGCPRunnerInitScript = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/gcp/init.sh"

// BuildInstallerSDKConfig renders an install's stack configuration: runner details,
// role permissions, roles, input values, and secrets. Serves the read-only config
// endpoint the Terraform provider reads.
func (h *Helpers) BuildInstallerSDKConfig(ctx context.Context, installID string) (*app.InstallerSDKConfig, error) {
	var install app.Install
	if res := h.db.WithContext(ctx).
		Preload("AWSAccount").
		// Newest row only: AfterQuery promotes it to CurrentInstallInputs, which the
		// input values and cluster_name below both read.
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC")).Limit(1)
		}).
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
		runnerAPIURL = h.cfg.RunnerAPIURL
	}
	if install.RunnerID == "" {
		return nil, fmt.Errorf("install %s has no runner — cannot build SDK config", installID)
	}
	if runnerAPIURL == "" {
		return nil, fmt.Errorf("install %s: runner_api_url empty (set RunnerGroupSettings.RunnerAPIURL for the install's runner group)", installID)
	}

	// GetFullAppConfig preloads PermissionsConfig, BreakGlassConfig,
	// InputConfig, SecretsConfig — same data the TF renderer walks.
	appCfg, err := h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, true)
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
	installState, err := h.GetInstallState(ctx, installID, false, false)
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

	// Real values, not names: the read is authenticated. Current value per
	// customer-source input, falling back to the app input's default.
	var currentInputs map[string]*string
	if install.CurrentInstallInputs != nil {
		currentInputs = install.CurrentInstallInputs.Values
	}

	var installInputs map[string]string
	var requiredInputs []string
	var sensitiveInputs []string
	for _, in := range appCfg.InputConfig.AppInputs {
		if in.Source != app.AppInputSourceCustomer {
			continue
		}
		if installInputs == nil {
			installInputs = map[string]string{}
		}

		value := in.Default
		if v, ok := currentInputs[in.Name]; ok && v != nil && *v != "" {
			value = *v
		}
		installInputs[in.Name] = value

		if in.Required {
			requiredInputs = append(requiredInputs, in.Name)
		}
		// install_inputs has no per-key metadata, so sensitivity travels as a
		// sibling name list, like required_inputs.
		if in.Sensitive {
			sensitiveInputs = append(sensitiveInputs, in.Name)
		}
	}

	// Auto-generated secrets are the stack's to mint, the rest the customer's to
	// supply. Vendor-owned values live on AppSecret, which this never reads.
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
		SensitiveInputs:     sensitiveInputs,
		AutoGenerateSecrets: autoGen,
		Secrets:             secrets,
	}

	// Runner machine/instance type from the app runner config, falling back to
	// the platform default — mirrors generate_install_stack_version's classic
	// tfvars path so both flows resolve the same type.
	instanceType := appCfg.RunnerConfig.InstanceType
	if instanceType == "" {
		instanceType = app.DefaultInstanceTypeForPlatform(appCfg.RunnerConfig.CloudPlatform)
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
		if h.cfg.RunnerDefaultSupportIAMRole != "" {
			supportARNs = []string{h.cfg.RunnerDefaultSupportIAMRole}
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
			Region:            install.AWSAccount.Region,
			ClusterName:       clusterName,
			RunnerMachineType: instanceType,

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
		token, err := h.runnersHelpers.CreateToken(ctx, install.RunnerID)
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
			RunnerMachineType:   instanceType,

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

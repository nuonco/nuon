package helpers

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	awsstacks "github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/aws"
	azurestacks "github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/azure"
	gcpstacks "github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/gcp"
)

const defaultGCPRunnerInitScript = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/gcp/init.sh"

func (h *Helpers) BuildInstallerSDKConfig(ctx context.Context, installID string) (*app.InstallerSDKConfig, error) {
	var install app.Install
	if res := h.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("GCPAccount").
		Preload("AzureAccount").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC")).Limit(1)
		}).
		Preload("RunnerGroup.Runners").
		Preload("RunnerGroup.Settings").
		Where("id = ?", installID).
		First(&install); res.Error != nil {
		return nil, fmt.Errorf("load install: %w", res.Error)
	}

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

	appCfg, err := h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, true)
	if err != nil {
		return nil, fmt.Errorf("load full app config: %w", err)
	}
	if appCfg == nil {
		return nil, fmt.Errorf("install %s: app config not found", installID)
	}

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

	app.ApplyInstallStackOverrides(&install, &appCfg.StackConfig)

	var currentInputs map[string]*string
	if install.CurrentInstallInputs != nil {
		currentInputs = install.CurrentInstallInputs.Values
	}

	var installInputs map[string]string
	var requiredInputs []string
	var sensitiveInputs []string
	customerInputNames := make(map[string]struct{})
	for _, in := range appCfg.InputConfig.AppInputs {
		if in.Source != app.AppInputSourceCustomer {
			continue
		}
		customerInputNames[in.Name] = struct{}{}
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
		if in.Sensitive {
			sensitiveInputs = append(sensitiveInputs, in.Name)
		}
	}

	customStacks, err := buildInstallerSDKCustomStacks(appCfg.StackConfig.CustomNestedStacks, stateData, customerInputNames)
	if err != nil {
		return nil, fmt.Errorf("build custom nested stacks: %w", err)
	}

	// why: Auto-generated secrets are the stack's to mint, the rest the customer's to
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
		CustomStacks:        customStacks,
	}

	var latestVersion app.InstallStackVersion
	res := h.db.WithContext(ctx).
		Where(app.InstallStackVersion{InstallID: install.ID}).
		Where("status->>'status' IN ?", app.InstallStackVersionTemplateReadyStatuses).
		Order("created_at DESC").
		Limit(1).
		First(&latestVersion)
	switch {
	case res.Error == nil:
		cfg.StackVersionID = latestVersion.ID
	case errors.Is(res.Error, gorm.ErrRecordNotFound):
	default:
		return nil, fmt.Errorf("load latest install stack version: %w", res.Error)
	}

	if (appCfg.RunnerConfig.Type == app.AppRunnerTypeAWS || appCfg.RunnerConfig.Type == app.AppRunnerTypeAzure) && len(customStacks) > 0 && latestVersion.ID != "" {
		cfg.CustomStacksTemplateURL = latestVersion.CustomStacksTemplateURL
		for i := range cfg.CustomStacks {
			cfg.CustomStacks[i].Outputs = latestVersion.CustomStacksOutputMap[cfg.CustomStacks[i].Name]
			cfg.CustomStacks[i].InputParameters = customerInputParameters(
				latestVersion.CustomStacksInputParametersMap[cfg.CustomStacks[i].Name],
				customerInputNames,
			)
		}
	}

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

		var supportARNs []string
		if h.cfg.RunnerDefaultSupportIAMRole != "" {
			supportARNs = []string{h.cfg.RunnerDefaultSupportIAMRole}
		}

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

			VPCNestedTemplateURL:    appCfg.StackConfig.VPCNestedTemplateURL,
			RunnerNestedTemplateURL: appCfg.StackConfig.RunnerNestedTemplateURL,

			NuonSupportIAMRoleARNs: supportARNs,

			ProvisionManagedPolicyARNs:      provMPAs,
			ProvisionInlinePolicyDocument:   provDoc,
			MaintenanceManagedPolicyARNs:    maintMPAs,
			MaintenanceInlinePolicyDocument: maintDoc,
			DeprovisionManagedPolicyARNs:    deprovMPAs,
			DeprovisionInlinePolicyDocument: deprovDoc,

			BreakGlassRoles: rolesToSDKConfigMap(breakGlass, false),
			CustomRoles:     rolesToSDKConfigMap(customRoles, true),
		}

	case app.AppRunnerTypeGCP:
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

		var gcpProjectID, gcpRegion string
		if install.GCPAccount != nil {
			gcpProjectID = install.GCPAccount.ProjectID
			gcpRegion = install.GCPAccount.Region
		}

		cfg.Cloud = "gcp"
		cfg.GCP = &app.InstallerSDKGCPConfig{
			ProjectID: gcpProjectID,
			Region:    gcpRegion,

			RunnerInitScriptURL: initScriptURL,
			RunnerAPIToken:      token.Token,
			RunnerMachineType:   instanceType,

			ProvisionPermissions:      prov.Permissions,
			ProvisionPredefinedRole:   prov.PredefinedRole,
			MaintenancePermissions:    maint.Permissions,
			MaintenancePredefinedRole: maint.PredefinedRole,
			DeprovisionPermissions:    deprov.Permissions,
			DeprovisionPredefinedRole: deprov.PredefinedRole,

			ProvisionPredefinedRoles:   prov.PredefinedRoles,
			MaintenancePredefinedRoles: maint.PredefinedRoles,
			DeprovisionPredefinedRoles: deprov.PredefinedRoles,

			ProvisionPolicies:   prov.Policies,
			MaintenancePolicies: maint.Policies,
			DeprovisionPolicies: deprov.Policies,

			BreakGlassRoles: gcpRolesToSDKMap(breakGlass, false),
			CustomRoles:     gcpRolesToSDKMap(customRoles, true),
		}

	case app.AppRunnerTypeAzure:
		if install.AzureAccount == nil || install.AzureAccount.Location == "" {
			return nil, fmt.Errorf("install %s has no Azure location; azure SDK provisioner requires it", installID)
		}

		prov, maint, deprov := azurestacks.ExtractAzureStandardRolesRaw(appCfg)
		breakGlass := azurestacks.ExtractAzureRolesRaw(appCfg.BreakGlassConfig.Roles)
		customRoles := azurestacks.ExtractAzureRolesRaw(appCfg.PermissionsConfig.CustomRoles)

		cfg.Cloud = "azure"
		cfg.Azure = &app.InstallerSDKAzureConfig{
			Location:             install.AzureAccount.Location,
			SubscriptionID:       install.AzureAccount.SubscriptionID,
			SubscriptionTenantID: install.AzureAccount.SubscriptionTenantID,

			VPCNestedTemplateURL:    appCfg.StackConfig.VPCNestedTemplateURL,
			RunnerNestedTemplateURL: appCfg.StackConfig.RunnerNestedTemplateURL,

			RunnerVMSize:      instanceType,
			ContainerImageURL: install.RunnerGroup.Settings.ContainerImageURL,
			ContainerImageTag: install.RunnerGroup.Settings.ContainerImageTag,

			ProvisionActions:        prov.Actions,
			ProvisionBuiltInRoles:   prov.BuiltInRoles,
			MaintenanceActions:      maint.Actions,
			MaintenanceBuiltInRoles: maint.BuiltInRoles,
			DeprovisionActions:      deprov.Actions,
			DeprovisionBuiltInRoles: deprov.BuiltInRoles,

			BreakGlassRoles: azureRolesToSDKMap(breakGlass, false),
			CustomRoles:     azureRolesToSDKMap(customRoles, true),
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

func buildInstallerSDKCustomStacks(stacks []config.CustomNestedStack, stateData map[string]any, customerInputNames map[string]struct{}) ([]app.InstallerSDKCustomStack, error) {
	if len(stacks) == 0 {
		return nil, nil
	}

	inputParameters := make(map[string]map[string]string, len(stacks))
	for _, stack := range stacks {
		for parameterName, value := range stack.Parameters {
			inputName, err := config.ParseInstallInputReference(value)
			if err != nil {
				continue
			}
			if _, ok := customerInputNames[inputName]; !ok {
				continue
			}
			if inputParameters[stack.Name] == nil {
				inputParameters[stack.Name] = make(map[string]string)
			}
			inputParameters[stack.Name][parameterName] = inputName
		}
	}

	if err := config.RenderCustomNestedStackParameters(stacks, stateData); err != nil {
		return nil, fmt.Errorf("render custom nested stack parameters: %w", err)
	}

	rendered := make([]config.CustomNestedStack, len(stacks))
	copy(rendered, stacks)
	sort.SliceStable(rendered, func(i, j int) bool {
		return rendered[i].Index < rendered[j].Index
	})

	out := make([]app.InstallerSDKCustomStack, len(rendered))
	for i, s := range rendered {
		out[i] = app.InstallerSDKCustomStack{
			Name:            s.Name,
			Index:           s.Index,
			Parameters:      s.Parameters,
			Module:          s.GCPModuleName(),
			InputParameters: inputParameters[s.Name],
		}
	}
	return out, nil
}

func customerInputParameters(inputParameters map[string]string, customerInputNames map[string]struct{}) map[string]string {
	var out map[string]string
	for parameterName, inputName := range inputParameters {
		if _, ok := customerInputNames[inputName]; !ok {
			continue
		}
		if out == nil {
			out = make(map[string]string)
		}
		out[parameterName] = inputName
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
			Permissions:     r.Permissions,
			PredefinedRole:  r.PredefinedRole,
			PredefinedRoles: r.PredefinedRoles,
			Enabled:         enabled,
			Policies:        r.Policies,
		}
	}
	return out
}

func azureRolesToSDKMap(rs []azurestacks.AzureRoleRaw, enabled bool) map[string]app.InstallerSDKAzureRole {
	if len(rs) == 0 {
		return nil
	}
	out := make(map[string]app.InstallerSDKAzureRole, len(rs))
	for _, r := range rs {
		out[r.Name] = app.InstallerSDKAzureRole{
			Actions:      r.Actions,
			BuiltInRoles: r.BuiltInRoles,
			Enabled:      enabled,
		}
	}
	return out
}

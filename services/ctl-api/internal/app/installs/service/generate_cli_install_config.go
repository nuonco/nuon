package service

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"github.com/pelletier/go-toml/v2"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						GenerateCLIInstallConfig
// @Summary				generate an install config to be used with CLI
// @Description.markdown	generate_cli_install_config.md
// @Param					install_id		path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				application/octet-stream
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{file}	config.Install
// @Router					/v1/installs/{install_id}/generate-cli-install-config [get]
func (s *service) GenerateCLIInstallConfig(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	installCfg, err := s.genCLIInstallConfig(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("error generating config from current state: %w", err))
		return
	}

	var response bytes.Buffer
	enc := toml.NewEncoder(&response)

	err = enc.Encode(installCfg)
	if err != nil {
		ctx.Error(fmt.Errorf("error encoding config: %w", err))
		return
	}

	output := strings.Replace(response.String(),
		"approval_option = ",
		"# Valid options: 'prompt' (default, requires manual approval) or 'approve-all' (automatic approval)\napproval_option = ",
		1)

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.toml\"", strcase.ToSnake(installCfg.Name)))
	ctx.Data(http.StatusOK, "application/octet-stream", []byte(output))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (s *service) genCLIInstallConfig(ctx context.Context, installID string) (*config.Install, error) {
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install %s: %w", installID, err)
	}

	installLabels := make(map[string]string, len(install.Labels)+len(install.LabelTemplates))
	for k, v := range install.Labels {
		installLabels[k] = v
	}
	for k, v := range install.LabelTemplates {
		installLabels[k] = v
	}
	for key := range install.AppDefaultLabels {
		delete(installLabels, key)
	}

	installCfg := config.Install{Name: install.Name, Labels: installLabels}
	if install.AppBranch != nil {
		installCfg.AppBranch = install.AppBranch.Name
	}
	installCfg.AppBranchGroup = install.AppBranchGroup

	if install.AWSAccount != nil {
		installCfg.AWSAccount = &config.AWSAccount{
			Region:    install.AWSAccount.Region,
			AccountID: install.CloudPlatformMetadata.TargetAccountID,
		}
	}
	if install.AzureAccount != nil {
		installCfg.AzureAccount = &config.AzureAccount{
			Location: install.AzureAccount.Location,
			SubscriptionID: firstNonEmpty(
				install.CloudPlatformMetadata.TargetSubscriptionID,
				install.AzureAccount.SubscriptionID,
			),
		}
	}
	if install.GCPAccount != nil {
		installCfg.GCPAccount = &config.GCPAccount{
			ProjectID: firstNonEmpty(
				install.CloudPlatformMetadata.TargetProjectID,
				install.GCPAccount.ProjectID,
			),
			Region: install.GCPAccount.Region,
		}
	}

	installConfig, err := s.helpers.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("failed parsing approval option: %w", err)
	}

	if installConfig != nil {
		if installConfig.TelemetryEnabled != nil {
			installCfg.Telemetry = &config.InstallTelemetry{Enabled: installConfig.TelemetryEnabled}
		}
		approvalOpt := config.InstallApprovalOption(installConfig.ApprovalOption)
		switch approvalOpt {
		case config.InstallApprovalOptionApproveAll:
			installCfg.ApprovalOption = config.InstallApprovalOptionApproveAll
		default:
			installCfg.ApprovalOption = config.InstallApprovalOptionPrompt
		}

		so := &config.InstallStackOverrides{}
		if installConfig.VPCNestedTemplateURL != nil {
			so.VPCNestedTemplateURL = *installConfig.VPCNestedTemplateURL
		}
		if installConfig.RunnerNestedTemplateURL != nil {
			so.RunnerNestedTemplateURL = *installConfig.RunnerNestedTemplateURL
		}
		if len(installConfig.CustomNestedStacks) > 0 {
			so.CustomNestedStacks = installConfig.CustomNestedStacks
		}
		if so.HasOverrides() {
			installCfg.StackOverrides = so
		}
	}

	appInputCfg, err := s.helpers.GetPinnedAppInputConfig(ctx, install.AppID, install.AppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app input config for install %s: %w", installID, err)
	}

	installInputs, err := s.getLatestInstallInputs(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get inputs for install %s: %w", installID, err)
	}

	installCfg.InputGroups = s.buildInputGroupsFromInputs(appInputCfg.AppInputs, installInputs.Values, s.l)
	installCfg.Components = buildComponentOverridesFromInputs(installInputs.Values)
	installCfg.ComponentToggles = buildComponentTogglesFromInputs(installInputs.Values)

	return &installCfg, nil
}

func buildComponentTogglesFromInputs(installInputValues map[string]*string) map[string]bool {
	toggles := make(map[string]bool)
	for name, val := range installInputValues {
		kind, compName, ok := config.ParseComponentOverrideInputName(name)
		if !ok || kind != config.ComponentOverrideKindEnabled {
			continue
		}
		v := generics.FromPtrStr(val)
		if v == "" {
			continue
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			continue
		}
		toggles[compName] = b
	}
	if len(toggles) == 0 {
		return nil
	}
	return toggles
}

func buildComponentOverridesFromInputs(installInputValues map[string]*string) map[string]config.ComponentOverride {
	components := make(map[string]config.ComponentOverride)
	for name, val := range installInputValues {
		kind, compName, ok := config.ParseComponentOverrideInputName(name)
		if !ok {
			continue
		}
		if kind == config.ComponentOverrideKindEnabled {
			continue
		}
		v := generics.FromPtrStr(val)
		if v == "" {
			continue
		}
		override := components[compName]
		switch kind {
		case config.ComponentOverrideKindHelmValues:
			override.HelmValues = v
		case config.ComponentOverrideKindTFVars:
			override.TFVars = v
		}
		components[compName] = override
	}
	if len(components) == 0 {
		return nil
	}
	return components
}

func (s *service) buildInputGroupsFromInputs(appInputs []app.AppInput, installInputValues map[string]*string, logger *zap.Logger) []config.InputGroup {
	inputGroupsMap := make(map[string]*config.InputGroup)

	for _, inp := range appInputs {
		if config.IsComponentOverrideInputName(inp.Name) {
			continue
		}

		if inputGroupsMap[inp.AppInputGroup.Name] == nil {
			inputGroupsMap[inp.AppInputGroup.Name] = &config.InputGroup{
				Inputs: make(map[string]string),
			}
		}

		if inp.Sensitive {
			continue
		}

		val, ok := installInputValues[inp.Name]
		if !ok {
			logger.Error("input is not set when generating install config",
				zap.String("key", inp.Name),
			)

			if inp.Required {
				inputGroupsMap[inp.AppInputGroup.Name].Inputs[inp.Name] = ""
			}
		} else {
			inputGroupsMap[inp.AppInputGroup.Name].Inputs[inp.Name] = generics.FromPtrStr(val)
		}
	}

	inputGroupsNames := slices.Collect(maps.Keys(inputGroupsMap))
	slices.Sort(inputGroupsNames)

	var result []config.InputGroup
	for _, groupName := range inputGroupsNames {
		ig := inputGroupsMap[groupName]
		if len(ig.Inputs) > 0 {
			result = append(result, config.InputGroup{
				Group:  groupName,
				Inputs: ig.Inputs,
			})
		}
	}

	return result
}

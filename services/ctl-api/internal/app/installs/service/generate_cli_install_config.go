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

	// Add a comment above the approval_option field to document valid values
	output := strings.Replace(response.String(),
		"approval_option = ",
		"# Valid options: 'prompt' (default, requires manual approval) or 'approve-all' (automatic approval)\napproval_option = ",
		1)

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.toml\"", strcase.ToSnake(installCfg.Name)))
	ctx.Data(http.StatusOK, "application/octet-stream", []byte(output))
}

func (s *service) genCLIInstallConfig(ctx context.Context, installID string) (*config.Install, error) {
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install %s: %w", installID, err)
	}

	installCfg := config.Install{
		Name:   install.Name,
		Labels: map[string]string(install.Labels),
	}

	if install.AWSAccount != nil {
		installCfg.AWSAccount = &config.AWSAccount{
			Region: install.AWSAccount.Region,
		}
	}

	installConfig, err := s.helpers.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("failed parsing approval option: %w", err)
	}

	if installConfig != nil {
		// Normalize the approval option: "auto" and empty both map to "prompt" in the generated config.
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

	return &installCfg, nil
}

// buildComponentOverridesFromInputs reconstructs the install config's
// [components.<name>] sections from the reserved synthetic override inputs, so a
// generated config round-trips symmetrically with what the user authored.
// Empty override values are omitted.
func buildComponentOverridesFromInputs(installInputValues map[string]*string) map[string]config.ComponentOverride {
	components := make(map[string]config.ComponentOverride)
	for name, val := range installInputValues {
		kind, compName, ok := config.ParseComponentOverrideInputName(name)
		if !ok {
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
		case config.ComponentOverrideKindEnabled:
			if b, err := strconv.ParseBool(v); err == nil {
				override.Enabled = generics.ToPtr(b)
			}
		}
		components[compName] = override
	}
	if len(components) == 0 {
		return nil
	}
	return components
}

// buildInputGroupsFromInputs constructs input groups from app inputs and install input values.
// it filters out sensitive inputs and only includes inputs that have values or are required.
// returns a sorted list of input groups with their corresponding inputs.
func (s *service) buildInputGroupsFromInputs(appInputs []app.AppInput, installInputValues map[string]*string, logger *zap.Logger) []config.InputGroup {
	// Build a map of input groups
	inputGroupsMap := make(map[string]*config.InputGroup)

	for _, inp := range appInputs {
		// Skip reserved per-component override inputs - they are emitted under
		// [components.<name>] instead of as flat inputs (see
		// buildComponentOverridesFromInputs).
		if config.IsComponentOverrideInputName(inp.Name) {
			continue
		}

		// Initialize input group if it doesn't exist
		if inputGroupsMap[inp.AppInputGroup.Name] == nil {
			inputGroupsMap[inp.AppInputGroup.Name] = &config.InputGroup{
				Inputs: make(map[string]string),
			}
		}

		// Skip sensitive inputs - they should not be included in the generated config
		if inp.Sensitive {
			continue
		}

		// Check if the input has a value in install inputs
		val, ok := installInputValues[inp.Name]
		if !ok {
			// Log error if input is not set
			logger.Error("input is not set when generating install config",
				zap.String("key", inp.Name),
			)

			// If input is required but not set, add it with empty string as placeholder
			if inp.Required {
				inputGroupsMap[inp.AppInputGroup.Name].Inputs[inp.Name] = ""
			}
			// If not required and not set, skip it entirely
		} else {
			// Add the input value to the group
			inputGroupsMap[inp.AppInputGroup.Name].Inputs[inp.Name] = generics.FromPtrStr(val)
		}
	}

	// Convert map to sorted slice
	inputGroupsNames := slices.Collect(maps.Keys(inputGroupsMap))
	slices.Sort(inputGroupsNames)

	var result []config.InputGroup
	for _, groupName := range inputGroupsNames {
		ig := inputGroupsMap[groupName]
		// Only include groups that have at least one input
		if len(ig.Inputs) > 0 {
			result = append(result, config.InputGroup{
				Group:  groupName,
				Inputs: ig.Inputs,
			})
		}
	}

	return result
}

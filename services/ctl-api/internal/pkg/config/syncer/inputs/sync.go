package inputs

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

var (
	interpolatedNameRegex = regexp.MustCompile(`^[a-z0-9_{}\.]*$`)
	validInputTypes       = map[string]bool{
		"bool":   true,
		"json":   true,
		"list":   true,
		"number": true,
		"string": true,
		"yaml":   true,
		"hcl":    true,
	}
)

// Sync creates the app input configuration with groups and inputs.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_input_config.go
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID, orgID string, state *sync.State) error {
	// Handle nil inputs config
	if cfg.Inputs == nil {
		inputCfg := app.AppInputConfig{
			AppConfigID:    appConfigID,
			OrgID:          orgID,
			AppID:          appID,
			AppInputGroups: []app.AppInputGroup{},
		}

		res := db.WithContext(ctx).Create(&inputCfg)
		if res.Error != nil {
			return sync.SyncInternalErr{
				Description: "unable to create empty app input config",
				Err:         fmt.Errorf("unable to create app input config: %w", res.Error),
			}
		}

		state.InputConfigID = inputCfg.ID
		return nil
	}

	// Validate inputs
	if err := validateInputs(cfg); err != nil {
		return sync.SyncErr{
			Resource:    "app-inputs",
			Description: fmt.Sprintf("validation failed: %v", err),
		}
	}

	// Create groups
	groups := make([]app.AppInputGroup, 0, len(cfg.Inputs.Groups))
	for idx, group := range cfg.Inputs.Groups {
		groups = append(groups, app.AppInputGroup{
			Name:        group.Name,
			Description: group.Description,
			DisplayName: group.DisplayName,
			Index:       idx,
		})
	}

	// Add synthetic override group if components need it
	if synthetic := config.SyntheticComponentOverrideInputs(cfg.Components); len(synthetic) > 0 {
		groups = append(groups, app.AppInputGroup{
			Name:        config.ComponentOverrideInputGroup,
			Description: "Reserved group for per-component install-level overrides (Helm values / Terraform vars).",
			DisplayName: "Component overrides",
			Index:       config.ComponentOverrideInputGroupIndex,
		})
	}

	inputCfg := app.AppInputConfig{
		AppConfigID:    appConfigID,
		OrgID:          orgID,
		AppID:          appID,
		AppInputGroups: groups,
	}

	res := db.WithContext(ctx).Create(&inputCfg)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app input config",
			Err:         fmt.Errorf("unable to create app input groups: %w", res.Error),
		}
	}

	// Create inputs
	var allInputs []app.AppInput

	if len(cfg.Inputs.Inputs) > 0 {
		for idx, input := range cfg.Inputs.Inputs {
			var groupID string
			for _, group := range inputCfg.AppInputGroups {
				if group.Name == input.Group {
					groupID = group.ID
					break
				}
			}

			source := app.AppInputSourceVendor
			if input.UserConfigurable {
				source = app.AppInputSourceCustomer
			}

			if input.Type == "json" && input.Default != nil {
				defaultStr := fmt.Sprintf("%v", input.Default)
				if defaultStr != "" && !json.Valid([]byte(defaultStr)) {
					return sync.SyncErr{
						Resource:    "app-inputs",
						Description: fmt.Sprintf("input %s has invalid JSON default value", input.Name),
					}
				}
			}

			var defaultVal string
			if input.Default != nil {
				defaultVal = fmt.Sprintf("%v", input.Default)
			}

			inputType := generics.ValOrDefault(input.Type, "string")

			allInputs = append(allInputs, app.AppInput{
				OrgID:            inputCfg.OrgID,
				AppInputConfigID: inputCfg.ID,
				AppInputGroupID:  groupID,
				Name:             input.Name,
				Description:      input.Description,
				DisplayName:      input.DisplayName,
				Required:         input.Required,
				Default:          defaultVal,
				Sensitive:        input.Sensitive,
				Type:             app.AppInputType(inputType),
				Index:            idx,
				Source:           source,
			})
		}
	}

	// Add synthetic component override inputs
	syntheticInputs := buildComponentOverrideInputs(cfg, &inputCfg)
	allInputs = append(allInputs, syntheticInputs...)

	if len(allInputs) > 0 {
		res := db.WithContext(ctx).Create(&allInputs)
		if res.Error != nil {
			return sync.SyncInternalErr{
				Description: "unable to create app inputs",
				Err:         fmt.Errorf("unable to create app inputs: %w", res.Error),
			}
		}
	}

	state.InputConfigID = inputCfg.ID
	return nil
}

func buildComponentOverrideInputs(cfg *config.AppConfig, inputCfg *app.AppInputConfig) []app.AppInput {
	synthetic := config.SyntheticComponentOverrideInputs(cfg.Components)
	if len(synthetic) == 0 {
		return nil
	}

	var overrideGroupID string
	for _, g := range inputCfg.AppInputGroups {
		if g.Name == config.ComponentOverrideInputGroup {
			overrideGroupID = g.ID
			break
		}
	}

	var inputs []app.AppInput
	for _, syn := range synthetic {
		desc, display := componentOverrideInputCopy(syn)
		inputs = append(inputs, app.AppInput{
			OrgID:            inputCfg.OrgID,
			AppInputConfigID: inputCfg.ID,
			AppInputGroupID:  overrideGroupID,
			Name:             syn.Name,
			Description:      desc,
			DisplayName:      display,
			Required:         false,
			Sensitive:        false,
			Type:             app.AppInputType(syn.Kind.InputType()),
			Index:            syn.Index,
			Source:           app.AppInputSourceVendor,
		})
	}
	return inputs
}

func componentOverrideInputCopy(syn config.SyntheticOverrideInput) (description, displayName string) {
	switch syn.Kind {
	case config.ComponentOverrideKindHelmValues:
		return fmt.Sprintf("Install-level Helm values override for component %q (YAML, deep-merged over app config).", syn.Component),
			fmt.Sprintf("%s helm values", syn.Component)
	case config.ComponentOverrideKindTFVars:
		return fmt.Sprintf("Install-level Terraform vars override for component %q (.tfvars, highest precedence).", syn.Component),
			fmt.Sprintf("%s tf vars", syn.Component)
	default:
		return fmt.Sprintf("Install-level override for component %q.", syn.Component),
			fmt.Sprintf("%s override", syn.Component)
	}
}

// validateInputs validates input names, group references, and types.
// Duplicates validation logic from services/ctl-api/internal/app/apps/service/create_app_input_config.go
func validateInputs(cfg *config.AppConfig) error {
	// Build map of valid group names
	validGroups := make(map[string]bool)
	for _, group := range cfg.Inputs.Groups {
		validGroups[group.Name] = true
	}

	// Validate each input
	for _, input := range cfg.Inputs.Inputs {
		// Validate input name using interpolated_name pattern
		if input.Name != "" && !interpolatedNameRegex.MatchString(input.Name) {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid input name: %s", input.Name),
				Description: fmt.Sprintf("Input name '%s' must contain only lowercase letters, numbers, underscores, dots, and curly braces (for interpolation)", input.Name),
			}
		}

		// Validate input references a valid group
		if input.Group != "" && !validGroups[input.Group] {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid group reference: %s", input.Group),
				Description: fmt.Sprintf("Input '%s' references group '%s' which does not exist", input.Name, input.Group),
			}
		}

		// Validate input type is valid
		inputType := generics.ValOrDefault(input.Type, "string")
		if !validInputTypes[inputType] {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid input type: %s", inputType),
				Description: fmt.Sprintf("Input '%s' has invalid type '%s'. Valid types are: bool, json, list, number, string, yaml, hcl", input.Name, inputType),
			}
		}
	}

	return nil
}

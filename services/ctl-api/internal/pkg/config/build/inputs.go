package build

import (
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
)

var validInputTypes = []app.AppInputType{
	app.AppInputTypeBool,
	app.AppInputTypeJSON,
	app.AppInputTypeList,
	app.AppInputTypeNumber,
	app.AppInputTypeString,
	app.AppInputTypeYAML,
	app.AppInputTypeHCL,
}

type InputGroupInput struct {
	Name        string
	DisplayName string
	Description string
	Index       int
}

type AppInputInput struct {
	Name        string
	DisplayName string
	Description string
	Default     string
	Group       string
	Type        string
	Required    bool
	Sensitive   bool
	Index       int
	Source      app.AppInputSource
}

// InputsFromConfig also materializes the reserved per-component override
// inputs: they must exist as app inputs for the install-input system to accept
// the values the install config carries under [components.<name>].
func InputsFromConfig(cfg *config.AppConfig) ([]InputGroupInput, []AppInputInput) {
	groups := []InputGroupInput{}
	inputs := []AppInputInput{}

	if cfg.Inputs != nil {
		for idx, group := range cfg.Inputs.Groups {
			groups = append(groups, InputGroupInput{
				Name:        group.Name,
				DisplayName: group.DisplayName,
				Description: group.Description,
				Index:       idx,
			})
		}

		for idx, input := range cfg.Inputs.Inputs {
			source := app.AppInputSourceVendor
			if input.UserConfigurable {
				source = app.AppInputSourceCustomer
			}

			var defaultVal string
			if input.Default != nil {
				defaultVal = fmt.Sprintf("%v", input.Default)
			}

			inputs = append(inputs, AppInputInput{
				Name:        input.Name,
				DisplayName: input.DisplayName,
				Description: input.Description,
				Default:     defaultVal,
				Group:       input.Group,
				Type:        generics.ValOrDefault(input.Type, string(app.AppInputTypeString)),
				Required:    input.Required,
				Sensitive:   input.Sensitive,
				Index:       idx,
				Source:      source,
			})
		}
	}

	synthetic := config.SyntheticComponentOverrideInputs(cfg.Components)
	if len(synthetic) == 0 {
		return dedupeByName(groups, groupName), dedupeByName(inputs, inputName)
	}

	groups = append(groups, InputGroupInput{
		Name:        config.ComponentOverrideInputGroup,
		DisplayName: "Component overrides",
		Description: "Reserved group for per-component install-level overrides (Helm values / Terraform vars).",
		Index:       config.ComponentOverrideInputGroupIndex,
	})

	for _, syn := range synthetic {
		description, displayName := ComponentOverrideInputCopy(syn)
		inputs = append(inputs, AppInputInput{
			Name:        syn.Name,
			DisplayName: displayName,
			Description: description,
			Default:     syn.Default,
			Group:       config.ComponentOverrideInputGroup,
			Type:        syn.Kind.InputType(),
			Index:       syn.Index,
			Source:      app.AppInputSourceVendor,
		})
	}

	return dedupeByName(groups, groupName), dedupeByName(inputs, inputName)
}

func groupName(g InputGroupInput) string { return g.Name }
func inputName(i AppInputInput) string   { return i.Name }

// dedupeByName keeps the last declaration of each name. A config may declare the
// same input or group twice; the request types this replaced were name-keyed
// maps, so those configs sync today and must keep syncing.
func dedupeByName[T any](items []T, name func(T) string) []T {
	lastAt := make(map[string]int, len(items))
	for idx, item := range items {
		lastAt[name(item)] = idx
	}

	out := make([]T, 0, len(lastAt))
	for idx, item := range items {
		if lastAt[name(item)] == idx {
			out = append(out, item)
		}
	}
	return out
}

func ComponentOverrideInputCopy(syn config.SyntheticOverrideInput) (description, displayName string) {
	switch syn.Kind {
	case config.ComponentOverrideKindHelmValues:
		return fmt.Sprintf("Install-level Helm values override for component %q (YAML, deep-merged over app config).", syn.Component),
			fmt.Sprintf("%s helm values", syn.Component)
	case config.ComponentOverrideKindTFVars:
		return fmt.Sprintf("Install-level Terraform vars override for component %q (.tfvars, highest precedence).", syn.Component),
			fmt.Sprintf("%s tf vars", syn.Component)
	case config.ComponentOverrideKindEnabled:
		return fmt.Sprintf("Whether component %q is deployed on this install. Set to false to tear it down, true to deploy it.", syn.Component),
			fmt.Sprintf("%s enabled", syn.Component)
	default:
		return fmt.Sprintf("Install-level override for component %q.", syn.Component),
			fmt.Sprintf("%s override", syn.Component)
	}
}

// InputConfig builds the config and its groups; inputs attach separately since
// they need the group IDs assigned on insert.
func InputConfig(groups []InputGroupInput, appID, appConfigID, orgID string) *app.AppInputConfig {
	objs := make([]app.AppInputGroup, 0, len(groups))
	for _, group := range groups {
		objs = append(objs, app.AppInputGroup{
			Name:        group.Name,
			DisplayName: group.DisplayName,
			Description: group.Description,
			Index:       group.Index,
		})
	}

	return &app.AppInputConfig{
		AppID:          appID,
		AppConfigID:    appConfigID,
		OrgID:          orgID,
		AppInputGroups: objs,
	}
}

// AppInputs resolves each input's group name to the ID assigned on insert.
func AppInputs(inputs []AppInputInput, cfg *app.AppInputConfig) ([]app.AppInput, error) {
	groupIDByName := make(map[string]string, len(cfg.AppInputGroups))
	for _, group := range cfg.AppInputGroups {
		groupIDByName[group.Name] = group.ID
	}

	objs := make([]app.AppInput, 0, len(inputs))
	for _, input := range inputs {
		if err := validation.ValidateInterpolatedName(input.Name); err != nil {
			return nil, err
		}

		if input.Group != "" {
			if _, ok := groupIDByName[input.Group]; !ok {
				return nil, stderr.ErrUser{
					Err:         fmt.Errorf("invalid group reference: %s", input.Group),
					Description: fmt.Sprintf("Input '%s' references group '%s' which does not exist", input.Name, input.Group),
				}
			}
		}

		inputType := generics.ValOrDefault(input.Type, string(app.AppInputTypeString))
		if !generics.SliceContains(app.AppInputType(inputType), validInputTypes) {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("invalid input type: %s", inputType),
				Description: fmt.Sprintf("Input '%s' has invalid type '%s'. Valid types are: bool, json, list, number, string, yaml, hcl", input.Name, inputType),
			}
		}

		if inputType == string(app.AppInputTypeJSON) && input.Default != "" && !json.Valid([]byte(input.Default)) {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("input %s default value is not a valid JSON string", input.Name),
				Description: fmt.Sprintf("Input '%s' has an invalid JSON default value", input.Name),
			}
		}

		objs = append(objs, app.AppInput{
			OrgID:            cfg.OrgID,
			AppInputConfigID: cfg.ID,
			AppInputGroupID:  groupIDByName[input.Group],
			Name:             input.Name,
			DisplayName:      input.DisplayName,
			Description:      input.Description,
			Default:          input.Default,
			Required:         input.Required,
			Sensitive:        input.Sensitive,
			Type:             app.AppInputType(inputType),
			Index:            input.Index,
			Source:           generics.ValOrDefault(input.Source, app.AppInputSourceVendor),
		})
	}

	return objs, nil
}

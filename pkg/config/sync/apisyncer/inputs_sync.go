package apisyncer

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) getAppInputRequest() *models.ServiceCreateAppInputConfigRequest {
	groups := make(map[string]models.ServiceAppGroupRequest)
	inputs := make(map[string]models.ServiceAppInputRequest)

	if s.cfg.Inputs != nil {
		for idx, group := range s.cfg.Inputs.Groups {
			group := group
			newGroup := models.ServiceAppGroupRequest{}
			newGroup.Description = &group.Description
			newGroup.DisplayName = &group.DisplayName
			newGroup.Index = generics.ToPtr(int64(idx))

			groups[group.Name] = newGroup
		}

		for idx, input := range s.cfg.Inputs.Inputs {
			input := input

			var inputSource models.AppAppInputSource
			if input.UserConfigurable {
				inputSource = models.AppAppInputSourceCustomer
			} else {
				inputSource = models.AppAppInputSourceVendor
			}

			inp := models.ServiceAppInputRequest{
				Description: &input.Description,
				DisplayName: &input.DisplayName,
				Group:       &input.Group,
				Required:    input.Required,
				Sensitive:   input.Sensitive,
				Type:        generics.ValOrDefault(input.Type, "string"),
				Source:      inputSource,
				Index:       generics.ToPtr(int64(idx)),
			}
			if input.Default != nil {
				inp.Default = fmt.Sprintf("%v", input.Default)
			}
			inputs[input.Name] = inp
		}
	}

	// Materialize reserved synthetic inputs for per-component install-level
	// overrides (Helm values / Terraform vars). These MUST be declared as app
	// inputs so the install-input system accepts, stores, and surfaces them; the
	// install config carries values for them under [components.<name>].
	s.addComponentOverrideInputs(groups, inputs)

	return &models.ServiceCreateAppInputConfigRequest{
		AppConfigID: s.appConfigID,
		Groups:      groups,
		Inputs:      inputs,
	}
}

// addComponentOverrideInputs injects the reserved group and one synthetic vendor
// input per Helm/Terraform component into the app input config request. They are
// vendor-sourced (so `nuon install inputs update` can edit them), not required,
// and default to empty (an empty override is an exact no-op at deploy time).
func (s *syncer) addComponentOverrideInputs(
	groups map[string]models.ServiceAppGroupRequest,
	inputs map[string]models.ServiceAppInputRequest,
) {
	synthetic := config.SyntheticComponentOverrideInputs(s.cfg.Components)
	if len(synthetic) == 0 {
		return
	}

	groupDesc := "Reserved group for per-component install-level overrides (Helm values / Terraform vars)."
	groupDisplay := "Component overrides"
	groups[config.ComponentOverrideInputGroup] = models.ServiceAppGroupRequest{
		Description: &groupDesc,
		DisplayName: &groupDisplay,
		Index:       generics.ToPtr(int64(config.ComponentOverrideInputGroupIndex)),
	}

	for _, syn := range synthetic {
		syn := syn
		desc, display := componentOverrideInputCopy(syn)
		group := config.ComponentOverrideInputGroup
		inputs[syn.Name] = models.ServiceAppInputRequest{
			Description: &desc,
			DisplayName: &display,
			Group:       &group,
			Required:    false,
			Sensitive:   false,
			Type:        syn.Kind.InputType(),
			Source:      models.AppAppInputSourceVendor,
			Index:       generics.ToPtr(int64(syn.Index)),
			Default:     syn.Default,
		}
	}
}

func componentOverrideInputCopy(syn config.SyntheticOverrideInput) (description, displayName string) {
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

func (s *syncer) syncAppInput(ctx context.Context, resource string) error {
	req := s.getAppInputRequest()
	cfg, err := s.apiClient.CreateAppInputConfig(ctx, s.appID, req)
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	s.state.InputConfigID = cfg.ID
	return nil
}

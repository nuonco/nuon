package sync

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s sync) getAppInputRequest() *models.ServiceCreateAppInputConfigRequest {
	// zero out the inputs if they are nil
	if s.cfg.Inputs == nil {
		return &models.ServiceCreateAppInputConfigRequest{
			AppConfigID: s.appConfigID,
			Groups:      make(map[string]models.ServiceAppGroupRequest, 0),
			Inputs:      make(map[string]models.ServiceAppInputRequest, 0),
		}
	}

	groups := make(map[string]models.ServiceAppGroupRequest)
	for idx, group := range s.cfg.Inputs.Groups {
		group := group
		groups[group.Name] = models.ServiceAppGroupRequest{
			Description: &group.Description,
			DisplayName: &group.DisplayName,
		}
		newGroup := models.ServiceAppGroupRequest{}
		newGroup.Description = &group.Description
		newGroup.DisplayName = &group.DisplayName
		newGroup.Index = generics.ToPtr(int64(idx))

		groups[group.Name] = newGroup
	}

	inputs := make(map[string]models.ServiceAppInputRequest)
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
			Internal:    input.Internal,
			Type:        generics.ValOrDefault(input.Type, "string"),
			Source:      inputSource,
			Index:       generics.ToPtr(int64(idx)),
		}
		if input.Default != nil {
			inp.Default = fmt.Sprintf("%v", input.Default)
		}
		inputs[input.Name] = inp
	}

	return &models.ServiceCreateAppInputConfigRequest{
		AppConfigID: s.appConfigID,
		Groups:      groups,
		Inputs:      inputs,
	}
}

func (s sync) syncAppInput(ctx context.Context, resource string) error {
	req := s.getAppInputRequest()
	cfg, err := s.apiClient.CreateAppInputConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	s.state.InputConfigID = cfg.ID
	return nil
}

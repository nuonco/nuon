package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"
)

func (s sync) getAppInputRequest() *models.ServiceCreateAppInputConfigRequest {
	if s.cfg.Inputs == nil {
		return &models.ServiceCreateAppInputConfigRequest{
			Groups: make(map[string]models.ServiceAppGroupRequest, 0),
			Inputs: make(map[string]models.ServiceAppInputRequest, 0),
		}
	}

	groups := make(map[string]models.ServiceAppGroupRequest)
	for _, group := range s.cfg.Inputs.Groups {
		group := group
		groups[group.Name] = models.ServiceAppGroupRequest{
			Description: &group.Description,
			DisplayName: &group.DisplayName,
		}
		newGroup := models.ServiceAppGroupRequest{}
		newGroup.Description = &group.Description
		newGroup.DisplayName = &group.DisplayName
		groups[group.Name] = newGroup
	}

	inputs := make(map[string]models.ServiceAppInputRequest)
	for _, input := range s.cfg.Inputs.Inputs {
		input := input
		inputs[input.Name] = models.ServiceAppInputRequest{
			Default:	input.Default,
			Description: &input.Description,
			DisplayName: &input.DisplayName,
			Group:       &input.Group,
			Required:    input.Required,
			Sensitive:   input.Sensitive,
		}
	}

	return &models.ServiceCreateAppInputConfigRequest{
		Groups: groups,
		Inputs: inputs,
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

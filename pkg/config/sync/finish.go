package sync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon-go/models"
)

func (s *sync) finish(ctx context.Context) error {
	stateJSON, err := json.Marshal(s.state)
	if err != nil {
		return fmt.Errorf("unable to convert state to json: %w", err)
	}

	compIDs := make([]string, 0)
	for _, comp := range s.state.Components {
		compIDs = append(compIDs, comp.ID)
	}

	if _, err := s.apiClient.UpdateAppConfig(ctx, s.appID, s.state.CfgID, &models.ServiceUpdateAppConfigRequest{
		State:             string(stateJSON),
		Status:            models.AppAppConfigStatusActive,
		StatusDescription: "successfully synced config",
		ComponentIds:      compIDs,
	}); err != nil {
		return err
	}

	return nil
}

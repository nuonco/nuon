package sync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon-go/models"
)

func (s *sync) syncStep(ctx context.Context, step syncStep) error {
	stepErr := step.Method(ctx)
	if stepErr == nil {
		return nil
	}

	stateJSON, err := json.Marshal(s.state)
	if err != nil {
		return fmt.Errorf("unable to convert state to json: %w", err)
	}

	_, err = s.apiClient.UpdateAppConfig(ctx, s.appID, s.state.CfgID, &models.ServiceUpdateAppConfigRequest{
		State:             string(stateJSON),
		Status:            models.AppAppConfigStatusError,
		StatusDescription: fmt.Sprintf("error updating %s", step.Resource),
	})
	if err != nil {
		fmt.Println("unable to update app config after failure: %w", err)
	}

	return stepErr
}

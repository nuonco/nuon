package apisyncer

import (
	"context"
	"encoding/json"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) start(ctx context.Context) error {
	req := &models.ServiceCreateAppConfigRequest{
		Readme:      s.cfg.Readme,
		CliVersion:  s.cliVersion,
		AppBranchID: s.appBranchID,
		PlanOnly:    s.planOnly,
	}

	intermediateJSON, err := json.Marshal(s.cfg)
	if err == nil {
		req.IntermediateConfigJSON = string(intermediateJSON)
	}

	cfg, err := s.apiClient.CreateAppConfig(ctx, s.appID, req)
	if err != nil {
		return err
	}

	s.appConfigID = cfg.ID
	s.state.CfgID = cfg.ID
	return nil
}

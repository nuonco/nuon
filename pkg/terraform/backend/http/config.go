package http

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *http) ConfigFile(ctx context.Context) ([]byte, error) {
	cfg := HTTPBackendConfig{
		Address:       s.Config.APIEndpoint + "/v1/terraform-backend?token=" + s.Config.Token + "&workspace_id=" + s.Config.WorkspaceID,
		LockAddress:   s.Config.APIEndpoint + "/v1/terraform-workspaces/" + s.Config.WorkspaceID + "/lock?token=" + s.Config.Token,
		UnlockAddress: s.Config.APIEndpoint + "/v1/terraform-workspaces/" + s.Config.WorkspaceID + "/unlock?token=" + s.Config.Token,
		LockMethod:    "POST",
		UnlockMethod:  "POST",
	}

	byts, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create config file: %w", err)
	}

	return byts, nil
}

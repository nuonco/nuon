package http

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *http) ConfigFile(ctx context.Context) ([]byte, error) {
	base := s.localBase
	if base == "" {
		if s.Config.Token != "" {
			return nil, fmt.Errorf("terraform backend proxy is not started")
		}
		base = strings.TrimRight(s.Config.APIEndpoint, "/")
	}

	cfg := HTTPBackendConfig{
		Address:       base + "/v1/terraform-backend?" + backendQuery("workspace_id", s.Config.WorkspaceID, "job_id", s.Config.JobID),
		LockAddress:   base + "/v1/terraform-workspaces/" + s.Config.WorkspaceID + "/lock?" + backendQuery("job_id", s.Config.JobID),
		UnlockAddress: base + "/v1/terraform-workspaces/" + s.Config.WorkspaceID + "/unlock?" + backendQuery("job_id", s.Config.JobID),
		LockMethod:    "POST",
		UnlockMethod:  "POST",
	}

	byts, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create config file: %w", err)
	}

	return byts, nil
}

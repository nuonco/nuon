package sync

import (
	"context"
	"encoding/json"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
)

const (
	defaultStateVersion string = "v1"
)

type componentState struct {
	Name     string                  `json:"name"`
	ID       string                  `json:"id"`
	ConfigID string                  `json:"config_id"`
	Type     models.AppComponentType `json:"type"`
}

type state struct {
	Version string `json:"version"`

	CfgID           string           `json:"config_id"`
	AppID           string           `json:"app_id"`
	InstallerID     string           `json:"installer_id"`
	RunnerConfigID  string           `json:"runner_config_id"`
	SandboxConfigID string           `json:"sandbox_config_id"`
	InputConfigID   string           `json:"input_config_id"`
	ComponentIDs    []componentState `json:"components"`
}

func (s *sync) fetchState(ctx context.Context) error {
	cfg, err := s.apiClient.GetAppLatestConfig(ctx, s.appID)
	if err != nil {
		if nuon.IsNotFound(err) {
			s.prevState = &state{}
			return nil
		}

		return err
	}

	var prevState state

	// NOTE(jm): this is required to handle in-flight configs, that do not have a previous state.
	if cfg.State != "" {
		if err := json.Unmarshal([]byte(cfg.State), &prevState); err != nil {
			return err
		}
	}

	s.prevState = &prevState
	return nil
}

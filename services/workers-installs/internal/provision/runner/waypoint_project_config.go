package runner

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type projectWaypointConfig struct {
	Project string `json:"project" validate:"required"`
}

func (r projectWaypointConfig) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// getProjectWaypointConfig returns an empty, but valid config for a waypoint project
func getProjectWaypointConfig(installID string) ([]byte, error) {
	cfg := projectWaypointConfig{
		Project: installID,
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("unable to validate config: %w", err)
	}

	byts, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to json: %w", err)
	}

	return byts, nil
}

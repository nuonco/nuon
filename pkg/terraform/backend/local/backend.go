package local

import (
	"context"
	"encoding/json"
	"fmt"
)

type localBackendConfig struct {
	Path string `json:"path"`
}

func (l *local) ConfigFile(ctx context.Context) ([]byte, error) {
	cfg := localBackendConfig{
		Path: l.Fp,
	}

	byts, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create config file: %w", err)
	}

	return byts, nil
}

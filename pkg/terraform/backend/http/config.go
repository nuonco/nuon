package http

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *http) ConfigFile(ctx context.Context) ([]byte, error) {
	cfg := HTTPBackendConfig{
		Address:       s.Config.Address,
		LockAddress:   s.Config.LockAddress,
		UnlockAddress: s.Config.UnlockAddress,
		LockMethod:    s.Config.LockMethod,
		UnlockMethod:  s.Config.UnlockMethod,
	}

	byts, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create config file: %w", err)
	}

	return byts, nil
}

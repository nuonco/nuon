package static

import (
	"context"
	"encoding/json"
	"fmt"
)

func (v *vars) Init(context.Context) error {
	return nil
}

func (v *vars) GetEnv(context.Context) (map[string]string, error) {
	return v.EnvVars, nil
}

func (v *vars) GetFile(context.Context) ([]byte, error) {
	if v.FileVars == nil {
		return nil, nil
	}

	byts, err := json.Marshal(v.FileVars)
	if err != nil {
		return nil, fmt.Errorf("unable to create file vars: %w", err)
	}

	return byts, nil
}

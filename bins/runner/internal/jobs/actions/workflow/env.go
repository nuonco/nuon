package workflow

import (
	"context"
)

const (
	outputsFilename string = "outputs.json"

	outputsEnvVar string = "NUON_ACTIONS_OUTPUT_FILEPATH"
	rootEnvVar           = "NUON_ACTIONS_ROOT"
)

func (h *handler) getBuiltInEnv(ctx context.Context) (map[string]string, error) {
	env := map[string]string{
		outputsEnvVar: h.state.workspace.AbsPath(outputsFilename),
		rootEnvVar:    h.state.workspace.Root(),
	}

	return env, nil
}

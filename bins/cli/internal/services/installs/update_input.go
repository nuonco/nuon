package installs

import (
	"context"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) UpdateInput(ctx context.Context, installID string, inputs []string, deployDependents bool, printJSON bool) error {
	inputsMap := make(map[string]string)
	for _, kv := range inputs {
		kvT := strings.Split(kv, "=")
		inputsMap[kvT[0]] = kvT[1]
	}
	request := &models.ServiceUpdateInstallInputsRequest{
		Inputs:           inputsMap,
		DeployDependents: &deployDependents,
	}
	if config.Debug() {
		ui.PrintJSON(request)
	}
	installInput, err := s.api.UpdateInstallInputs(ctx, installID, request)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	ui.PrintJSON(installInput)
	return nil
}

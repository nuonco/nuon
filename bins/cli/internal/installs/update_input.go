package installs

import (
	"context"
	"strings"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) UpdateInput(ctx context.Context, installID string, inputs []string, printJSON bool) error {
	inputsMap := make(map[string]string)
	for _, kv := range inputs {
		kvT := strings.Split(kv, "=")
		inputsMap[kvT[0]] = kvT[1]
	}
	installInput, _, err := s.api.UpdateInstallInputs(ctx, installID, &models.ServiceUpdateInstallInputsRequest{
		Inputs: inputsMap,
	})
	if err != nil {
		return ui.PrintJSONError(err)
	}

	ui.PrintJSON(installInput)
	return nil
}

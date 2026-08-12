package installs

import (
	"context"
	"strings"

	"github.com/nuonco/nuon/pkg/cli/styles"
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
		if printJSON {
			return err
		}
		return err
	}

	if printJSON {
		ui.PrintJSON(installInput)
		return nil
	}

	values := redactedValues(installInput)
	view := ui.NewGetView()
	data := [][]string{{"", "VALUE"}}
	for k, v := range inputsMap {
		val, ok := values[k]
		if !ok {
			val = v
		}
		data = append(data, []string{styles.TextPrimary.Render(k), val})
	}
	view.Render(data)
	return nil
}

package helpers

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ValidateRunbookInputs verifies the supplied inputs against the runbook config's
// declared inputs: no unknown names, and every required input has a non-empty value.
func (s *Helpers) ValidateRunbookInputs(rbConfig *app.RunbookConfig, inputs map[string]*string) error {
	if rbConfig == nil || len(rbConfig.Inputs) == 0 {
		if len(inputs) > 0 {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid runbook inputs provided"),
				Description: "inputs provided to the runbook run that are not defined on the runbook",
			}
		}
		return nil
	}

	inputNames := map[string]struct{}{}
	for _, input := range rbConfig.Inputs {
		inputNames[input.Name] = struct{}{}
	}

	for name := range inputs {
		if _, ok := inputNames[name]; !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("input name %s does not exist in runbook inputs", name),
				Description: "input " + name + " provided to the runbook run does not exist in the runbook inputs",
			}
		}
	}

	for _, inp := range rbConfig.Inputs {
		if !inp.Required {
			continue
		}

		inputVal, ok := inputs[inp.Name]
		if !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("%s is a required input", inp.Name),
				Description: fmt.Sprintf("%s is required, please add a value for the input", inp.Name),
			}
		}
		if inputVal == nil || len(*inputVal) < 1 {
			return stderr.ErrUser{
				Err:         fmt.Errorf("%s must be non-empty", inp.Name),
				Description: fmt.Sprintf("%s is required, please add a non-empty value for the input", inp.Name),
			}
		}
	}

	return nil
}

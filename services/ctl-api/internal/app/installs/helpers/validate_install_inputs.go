package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (s *Helpers) ValidateInstallInputs(ctx context.Context, appInputCfg *app.AppInputConfig, inputs map[string]*string) error {
	if appInputCfg == nil {
		if len(inputs) > 0 {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid install inputs provided"),
				Description: "inputs provided on install, that are not defined on the app",
			}
		}

		return nil
	}

	// verify all of the inputs are defined in the app input config
	appInputNames := map[string]struct{}{}
	inputTypes := map[string]app.AppInputType{}
	for _, input := range appInputCfg.AppInputs {
		appInputNames[input.Name] = struct{}{}
		inputTypes[input.Name] = input.Type
	}

	for name, val := range inputs {
		if _, ok := appInputNames[name]; !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("input name %s does not exist in app inputs", name),
				Description: "input " + name + " defined for install does not exist in the app inputs",
			}
		}

		// structured inputs (yaml/hcl) must parse cleanly so a malformed
		// override is rejected here rather than failing mid-deploy.
		if val != nil {
			if err := config.ValidateInputValueSyntax(string(inputTypes[name]), *val); err != nil {
				return stderr.ErrUser{
					Err:         err,
					Description: fmt.Sprintf("input %s has an invalid value: %s", name, err.Error()),
				}
			}
		}
	}

	// verify all of the inputs are set on the current sandbox config
	for _, inp := range appInputCfg.AppInputs {
		if !inp.Required ||
			inp.Source == app.AppInputSourceCustomer {
			continue
		}

		inputVal, ok := inputs[inp.Name]
		if !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("%s is a required input", inp.Name),
				Description: fmt.Sprintf("%s is required, please add a value value for the input", inp.Name),
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

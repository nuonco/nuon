package vars

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (v *varsValidator) getTemplate(ctx context.Context) (map[string]interface{}, error) {
	obj := generics.GetFakeObj[intermediate]()

	inps := v.getInputs()
	obj.Install.Inputs = inps

	// add components
	compOut := v.getComponents()
	obj.Components = compOut

	// install stack
	stackOut := v.getInstallStack()
	obj.InstallStack = stackOut

	// add the vcs config into the fake data
	if !v.ignoreSandboxOutputs {
		out, err := v.getSandboxOutputs(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get template sandbox outputs")
		}

		obj.Install.Sandbox.Outputs = out
		obj.Sandbox.Outputs = out
	}

	data, err := v.toInterfaceMap(obj)
	if err != nil {
		return nil, errors.Wrap(err, "internal error generating fake data")
	}

	return data, nil
}

func (s *varsValidator) toInterfaceMap(i intermediate) (map[string]interface{}, error) {
	byts, err := json.Marshal(map[string]interface{}{
		"nuon": i,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to convert to json: %w", err)
	}

	var tmplData map[string]interface{}
	if err := json.Unmarshal(byts, &tmplData); err != nil {
		return nil, fmt.Errorf("unable to convert from json to int map: %w", err)
	}

	return tmplData, nil
}

package protos

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	variablesv1 "github.com/powertoolsdev/mono/pkg/types/components/variables/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) toInstallInputs(install app.Install) (*variablesv1.InstallInputs, error) {
	if len(install.InstallInputs) < 1 || install.CurrentInstallInputs == nil {
		return &variablesv1.InstallInputs{}, nil
	}

	inputs := make([]*variablesv1.InstallInput, 0)
	appInputs := install.CurrentInstallInputs.AppInputConfig
	for _, input := range appInputs.AppInputs {
		installInput, ok := install.CurrentInstallInputs.Values[input.Name]
		if !ok || installInput == nil || *installInput == "" {
			installInput = generics.ToPtr(input.Default)
		}

		if (installInput == nil || *installInput == "") && input.Required {
			return nil, fmt.Errorf("install is missing required input %s", input.Name)
		}

		inputs = append(inputs, &variablesv1.InstallInput{
			Name:  input.Name,
			Value: generics.FromPtrStr(installInput),
		})
	}

	return &variablesv1.InstallInputs{
		InstallInputs: inputs,
	}, nil
}

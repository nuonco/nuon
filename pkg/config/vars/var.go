package vars

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config"
)

type ValidateVarsParams struct {
	Cfg  *config.AppConfig
	Vars []string

	IgnoreComponent      string
	IgnoreSandboxOutputs bool
}

type varsValidator struct {
	cfg *config.AppConfig

	ignoreComponent      string
	ignoreSandboxOutputs bool
}

func ValidateVars(ctx context.Context, params ValidateVarsParams) error {
	v := &varsValidator{
		cfg:                  params.Cfg,
		ignoreComponent:      params.IgnoreComponent,
		ignoreSandboxOutputs: params.IgnoreSandboxOutputs,
	}

	tmplData, err := v.getTemplate(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get template data")
	}

	for _, inputVar := range params.Vars {
		if err := v.validateVar(inputVar, tmplData); err != nil {
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("var %s was not valid", inputVar))
			}
		}
	}

	return nil
}

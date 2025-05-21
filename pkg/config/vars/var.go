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

	fakeState, err := v.GetFakeState(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get fake state")
	}

	fakeStateMap, err := fakeState.AsMap()
	if err != nil {
		return errors.Wrap(err, "unable to convert fake state to map")
	}

	for _, inputVar := range params.Vars {
		if err := v.validateVarV2(inputVar, fakeStateMap); err != nil {
			return errors.Wrap(err, fmt.Sprintf("var %s was not valid", inputVar))
		}
	}

	return nil
}

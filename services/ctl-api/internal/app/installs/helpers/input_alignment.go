package helpers

import (
	"slices"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/inputstate"
)

func InputNames(references []refs.Ref) []string {
	names := make([]string, 0)
	for _, ref := range references {
		if ref.Type == refs.RefTypeInputs || ref.Type == refs.RefTypeInstallInputs {
			names = append(names, ref.Name)
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

func ComponentInputNames(cfg *app.ComponentConfigConnection, inputs *state.InputsState) []string {
	names := InputNames(cfg.Refs)
	name := cfg.Component.Name
	if name == "" {
		name = cfg.ComponentName
	}
	if cfg.IsToggleable() {
		names = append(names, config.EnabledOverrideInputName(name))
	}
	switch {
	case cfg.Type == app.ComponentTypeHelmChart || cfg.Component.Type == app.ComponentTypeHelmChart:
		names = append(names, config.HelmValuesOverrideInputName(name))
	case cfg.Type == app.ComponentTypeTerraformModule || cfg.Component.Type == app.ComponentTypeTerraformModule:
		names = append(names, config.TFVarsOverrideInputName(name))
	}
	if inputs != nil {
		for _, override := range []string{config.HelmValuesOverrideInputName(name), config.TFVarsOverrideInputName(name)} {
			if raw := inputs.Inputs[override]; raw != "" {
				names = append(names, override)
				names = append(names, InputNames(refs.ParseFieldRefs(raw))...)
			}
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

func InputsDigest(names []string, inputs *state.InputsState) string {
	return inputstate.Digest(names, inputs)
}

func FillInputDefaults(st *state.State, cfg *app.AppConfig) {
	inputstate.FillDefaults(st, cfg)
}

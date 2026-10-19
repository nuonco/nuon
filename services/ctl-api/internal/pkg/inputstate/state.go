package inputstate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func Resolve(inputs *app.InstallInputs, cfg *app.AppConfig, redacted bool) *state.InputsState {
	if inputs == nil {
		return nil
	}
	values := inputs.Values
	if redacted {
		values = inputs.ValuesRedacted
	}
	if len(values) == 0 && len(cfg.InputConfig.AppInputs) == 0 {
		return nil
	}
	resolved := state.NewInputsState()
	for _, input := range cfg.InputConfig.AppInputs {
		value := generics.FromPtrStr(values[input.Name])
		if value == "" {
			value = input.Default
		}
		resolved.Inputs[input.Name] = value
	}
	return resolved
}

func Digest(names []string, inputs *state.InputsState) string {
	values := make(map[string]string, len(names))
	for _, name := range names {
		values[name] = ""
		if inputs != nil {
			values[name] = inputs.Inputs[name]
		}
	}
	data, _ := json.Marshal(values)
	return fmt.Sprintf("v1:%x", sha256.Sum256(data))
}

func FillDefaults(st *state.State, cfg *app.AppConfig) {
	if st == nil || cfg == nil {
		return
	}
	for _, input := range cfg.InputConfig.AppInputs {
		if input.Default == "" {
			continue
		}
		if st.Inputs == nil {
			st.Inputs = state.NewInputsState()
			st.Inputs.Populated = true
		}
		if st.Inputs.Inputs[input.Name] != "" {
			continue
		}
		st.Inputs.Inputs[input.Name] = input.Default
		if st.Install != nil {
			if st.Install.Inputs == nil {
				st.Install.Inputs = make(map[string]string)
			}
			st.Install.Inputs[input.Name] = input.Default
		}
	}
}

package helpers

import (
	"encoding/json"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ChangedInputValue struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type ChangedInputsResult struct {
	Names             []string
	ChangedValuesJSON string
	ChangedValues     map[string]ChangedInputValue
}

func ComputeChangedInputs(
	oldValues map[string]*string,
	newValues map[string]*string,
	appInputs []app.AppInput,
) (*ChangedInputsResult, error) {
	sensitiveInputs := make(map[string]bool)
	for _, input := range appInputs {
		if input.Sensitive {
			sensitiveInputs[input.Name] = true
		}
	}

	var names []string
	changedValues := make(map[string]ChangedInputValue)
	for k, newPtr := range newValues {
		oldPtr := oldValues[k]
		oldVal := generics.FromPtrStr(oldPtr)
		newVal := generics.FromPtrStr(newPtr)

		if oldVal != newVal {
			names = append(names, k)

			displayOld := oldVal
			displayNew := newVal
			if sensitiveInputs[k] {
				displayOld = "***"
				displayNew = "***"
			}
			changedValues[k] = ChangedInputValue{Old: displayOld, New: displayNew}
		}
	}

	valuesJSON := ""
	if len(changedValues) > 0 {
		b, err := json.Marshal(changedValues)
		if err != nil {
			return nil, err
		}
		valuesJSON = string(b)
	}

	return &ChangedInputsResult{
		Names:             names,
		ChangedValuesJSON: valuesJSON,
		ChangedValues:     changedValues,
	}, nil
}

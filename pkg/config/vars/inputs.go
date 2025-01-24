package vars

import (
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (v *varsValidator) getInputs() map[string]interface{} {
	inps := make(map[string]interface{}, 0)
	for _, inp := range v.cfg.Inputs.Inputs {
		inps[inp.Name] = generics.GetFakeObj[string]()
	}

	return inps
}

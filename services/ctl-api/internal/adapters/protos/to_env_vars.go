package protos

import variablesv1 "github.com/powertoolsdev/mono/pkg/types/components/variables/v1"

func (c *Adapter) toEnvVars(inputVals map[string]*string) *variablesv1.EnvVars {
	vals := make([]*variablesv1.EnvVar, 0)
	for k, v := range inputVals {
		if v == nil {
			continue
		}

		vals = append(vals, &variablesv1.EnvVar{
			Name:      k,
			Value:     *v,
			Sensitive: true,
		})
	}

	return &variablesv1.EnvVars{
		Env: vals,
	}
}

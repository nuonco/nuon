package static

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
)

func (v *vars) Init(context.Context) error {
	return nil
}

func (v *vars) GetEnv(context.Context) (map[string]string, error) {
	return v.EnvVars, nil
}

func (v *vars) GetFiles(context.Context) ([]variables.VarFile, error) {
	files := make([]variables.VarFile, 0)

	if v.FileVars != nil {
		byts, err := json.Marshal(v.FileVars)
		if err != nil {
			return nil, fmt.Errorf("unable to create file vars: %w", err)
		}

		files = append(files, variables.VarFile{
			Filename: "vars-0.json",
			Contents: byts,
		})
	}

	for idx, file := range v.Files {
		ext := "tfvars"
		if generics.IsJSONStr(file) {
			ext = "json"
		}
		fn := fmt.Sprintf("vars-%d.%s", idx+1, ext)

		files = append(files, variables.VarFile{
			Filename: fn,
			Contents: []byte(file),
		})
	}

	return files, nil
}

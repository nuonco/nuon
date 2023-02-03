package executor

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type outputter interface {
	Output(context.Context, ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error)
}

var _ outputter = (*tfexec.Terraform)(nil)

// Output runs terraform output for the current module
func (e *tfExecutor) Output(ctx context.Context) (map[string]interface{}, error) {
	m := map[string]interface{}{}

	out, err := e.outputter.Output(ctx)
	if err != nil {
		return m, err
	}

	for k, v := range out {
		var val interface{}
		err = json.Unmarshal(v.Value, &val)
		if err != nil {
			return m, err
		}
		m[k] = val
	}

	return m, err
}

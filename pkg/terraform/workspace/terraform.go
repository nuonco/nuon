package workspace

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func (w *workspace) Apply(ctx context.Context) (*tfjson.Plan, error) {
	return nil, nil
}

func (w *workspace) Destroy(ctx context.Context) (*tfjson.Plan, error) {
	return nil, nil
}

func (w *workspace) Plan(ctx context.Context) (*tfjson.Plan, error) {
	return nil, nil
}

func (w *workspace) Output(ctx context.Context) (map[string]tfexec.OutputMeta, error) {
	return nil, nil
}

func (w *workspace) Show(ctx context.Context) (*tfjson.State, error) {
	return nil, nil
}

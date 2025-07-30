package workspace

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"
)

func (w *workspace) ShowPlan(ctx context.Context, log hclog.Logger) (*tfjson.Plan, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.showPlan(ctx, client)
}

func (w *workspace) showPlan(ctx context.Context, client Terraform) (*tfjson.Plan, error) {
	// NOTE: takes the tf plan and returns a json serializable *struct
	out, err := client.ShowPlanFile(ctx, "tfplan")
	if err != nil {
		return nil, fmt.Errorf("unable to execute show: %w", err)
	}
	return out, nil
}

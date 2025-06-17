package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

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

// TODO: revisit and use a callback to write to local instead of writing to local directly
func (w *workspace) showPlan(ctx context.Context, client Terraform) (*tfjson.Plan, error) {
	out, err := client.ShowPlanFile(ctx, "tfplan")
	if err != nil {
		return nil, fmt.Errorf("unable to execute show: %w", err)
	}

	planJSON, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("unable to marshal plan to JSON: %w", err)
	}

	// TODO: this should be legible from the workspace root but something is wrong in
	// the local file writer callback pkg/terraform
	pathToPlan := path.Join(w.Root(), "plan.json")
	if err := os.WriteFile(pathToPlan, planJSON, 0644); err != nil {
		return nil, fmt.Errorf("unable to write plan to file: %w", err)
	}

	return out, nil
}

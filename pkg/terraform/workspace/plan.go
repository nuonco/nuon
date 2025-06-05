package workspace

import (
	"context"
	"fmt"
)

const (
	defaultPlanConfigFilename = "plan.json"
)

// WritePlan writes the Terraform plan JSON to a file in the workspace
func (w *workspace) WritePlan(ctx context.Context, plan string) error {
	// NOTE: the plan is expected to be a json tf plan

	// Create the plan.json file in the workspace directory
	planFilePath := w.root + "/" + defaultPlanConfigFilename

	// Write the JSON plan string to the file
	err := w.writeFile(planFilePath, []byte(plan), 0644)
	if err != nil {
		return fmt.Errorf("unable to write %s file: %w", defaultPlanConfigFilename, err)
	}

	return nil
}

package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
)

const (
	defaultPlanFilename = "tfplan"
)

// WritePlan writes the Terraform tfplan to a file called tfplan in the workspace
func (w *workspace) WriteTFPlan(ctx context.Context, log hclog.Logger) ([]byte, error) {
	// NOTE: the plan is expected to be an opaque format tfplan file (not human legible)

	// Create the plan.json file in the workspace directory
	planFilePath := filepath.Join(w.root, defaultPlanFilename)
	log.Debug("writing plan", zap.String("path", planFilePath), zap.Int("plan.bytes.count", len(w.PlanBytes)))
	// Write the JSON plan string to the file
	fd, err := os.Create(planFilePath)
	defer fd.Close()
	if err != nil {
		return []byte{}, fmt.Errorf("unable to create %s file: %w", defaultPlanFilename, err)
	}

	n, err := fd.Write(w.PlanBytes)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to write %s file: %w", defaultPlanFilename, err)
	}
	log.Debug("wrote plan", zap.String("path", planFilePath), zap.Int("plan.bytes.count", len(w.PlanBytes)), zap.Int("bytes-written", n))
	fd.Sync()

	return w.PlanBytes, nil
}

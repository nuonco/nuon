package workflow

import (
	"context"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/actions/outputs"
	"github.com/nuonco/nuon/pkg/generics"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

func (h *handler) outputsFP(cfg *models.AppActionWorkflowStepConfig) string {
	return filepath.Join(h.state.workspace.Root(), outputs.Filename(cfg.Idx))
}

func (h *handler) parseOutputs(ctx context.Context) (map[string]interface{}, error) {
	steps := make(map[string]any, 0)
	merged := make(map[string]interface{}, 0)

	// build the list of step configs to read outputs from
	var stepCfgs []*models.AppActionWorkflowStepConfig
	if h.state.workflowCfg != nil {
		stepCfgs = h.state.workflowCfg.Steps
	} else {
		// for adhoc actions, build configs from run steps
		for idx, step := range h.state.run.Steps {
			if step.AdhocConfig != nil {
				stepCfgs = append(stepCfgs, &models.AppActionWorkflowStepConfig{
					Idx:  int64(idx),
					Name: step.AdhocConfig.Name,
				})
			}
		}
	}

	if len(stepCfgs) == 0 {
		return merged, nil
	}

	for _, stepCfg := range stepCfgs {
		stepOutputs, err := outputs.ParseFile(h.outputsFP(stepCfg))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse outputs for step %s", stepCfg.Name)
		}

		merged = generics.MergeMap(merged, stepOutputs)
		steps[stepCfg.Name] = stepOutputs
	}

	merged["steps"] = steps
	return merged, nil
}

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, err
	}

	outs, err := h.parseOutputs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse outputs")
	}

	l.Debug("successfully parsed action workflow outputs", zap.Any("outputs", outs))

	return outs, nil
}

package workflow

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	kvDelimiter     string = "="
	jsonObjStart    string = "{"
	jsonArrayStart  string = "["
	outputsFilename string = "%d.nuon-outputs.json"
)

func (h *handler) outputsFP(cfg *models.AppActionWorkflowStepConfig) string {
	fn := fmt.Sprintf(outputsFilename, cfg.Idx)
	return filepath.Join(h.state.workspace.Root(), fn)
}

func (h *handler) parseOutputLine(ctx context.Context, str string) (map[string]interface{}, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, err
	}

	// look to parse the string as a json map
	if strings.HasPrefix(str, jsonObjStart) {
		var out map[string]interface{}
		if err := json.Unmarshal([]byte(str), &out); err != nil {
			return nil, errors.Wrap(err, "unable to parse as json")
		}

		return out, nil
	}

	// check to make sure it is not a json array
	if strings.HasPrefix(str, jsonArrayStart) {
		l.Error("outputs with top level json arrays are not yet supported")
		return nil, errors.New("outputs with top level json arrays are not supported yet")
	}

	// check to see if it is a key value
	pieces := strings.SplitN(str, kvDelimiter, 2)
	if len(pieces) == 2 {
		return map[string]interface{}{
			pieces[0]: pieces[1],
		}, nil
	}

	l.Error("output format not supported, must be one a json object or k=v string", zap.String("output", str))
	return nil, errors.New("unsupported outputs format")
}

func (h *handler) parseOutputs(ctx context.Context) (map[string]interface{}, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, err
	}

	steps := make(map[string]any, 0)
	outputs := make(map[string]interface{}, 0)

	for _, stepCfg := range h.state.workflowCfg.Steps {
		fh, err := os.Open(h.outputsFP(stepCfg))
		if err != nil {
			l.Error("error opening outputs file", zap.Error(err))
			return nil, errors.Wrap(err, "unable to get outputs file")
		}
		defer fh.Close()

		stepOutputs := make(map[string]interface{}, 0)

		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			line := scanner.Text()
			lineOutputs, err := h.parseOutputLine(ctx, line)
			if err != nil {
				l.Error("error parsing outputs line", zap.Error(err))
				return nil, errors.Wrap(err, "error parsing outputs")
			}

			stepOutputs = generics.MergeMap(stepOutputs, lineOutputs)
		}
		if err := scanner.Err(); err != nil {
			return nil, errors.Wrap(err, "unable to scan outputs file")
		}

		outputs = generics.MergeMap(outputs, stepOutputs)
		steps[stepCfg.Name] = stepOutputs
	}

	outputs["steps"] = steps
	return outputs, nil
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

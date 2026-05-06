package helm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/op"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
	"github.com/nuonco/nuon/pkg/plans"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("packaging chart", zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.state.arch.BasePath()})
	_, endPkg := op.Tool(ctx, "helm", "package_chart")
	packagePath, err := h.packageChart(l)
	endPkg(err)
	if err != nil {
		return fmt.Errorf("unable to get source files: %w", err)
	}
	l.Info("packaged chart", zapcore.Field{Key: "package_path", Type: zapcore.StringType, String: packagePath})

	l.Info("successfully packaged chart", zap.String("path", packagePath))
	h.state.packagePath = packagePath

	srcFiles, err := h.getSourceFiles()
	if err != nil {
		return errors.Wrap(err, "unable to get source files")
	}

	l.Info("packing chart into archive")
	if err := h.state.arch.Pack(ctx, l, srcFiles); err != nil {
		return fmt.Errorf("unable to pack archive with helm archive: %w", err)
	}

	l.Info("copying archive to destination")
	res, err := h.ociCopy.CopyFromStore(ctx,
		h.state.arch.Ref(),
		"latest",
		h.state.regCfg,
		h.state.resultTag,
	)
	if err != nil {
		h.writeErrorResult(ctx, "copy image", err)
		return err
	}

	l.Info("writing job result")
	resultReq := registry.ToAPIResult(res)
	h.appendPolicyInputToResult(ctx, l, resultReq)
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}
	return nil
}

func (h *handler) appendPolicyInputToResult(ctx context.Context, l *zap.Logger, resultReq *models.ServiceCreateRunnerJobExecutionResultRequest) {
	policyInput, policyErr := h.buildPolicyInput(ctx, l)
	if policyErr != nil {
		h.errRecorder.Record("build policy input", policyErr)
		return
	}
	if policyInput == nil {
		l.Debug("policy input missing or empty")
		return
	}

	l.Debug("policy input generated", zap.Int("policy_input_count", len(policyInput)))

	contentsJSON, err := json.Marshal(map[string]any{
		"policy_input": policyInput,
	})
	if err != nil {
		h.errRecorder.Record("marshal policy input", err)
		return
	}

	compressedContents, err := plans.CompressPlan(contentsJSON)
	if err != nil {
		h.errRecorder.Record("compress policy input", err)
		return
	}

	resultReq.ContentsCompressed = compressedContents
	l.Debug("policy input appended to job result", zap.Int("compressed_bytes", len(compressedContents)))
}

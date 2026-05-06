package helm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	pkgop "github.com/nuonco/nuon/bins/runner/internal/pkg/op"
	"github.com/nuonco/nuon/pkg/diff"
	"github.com/nuonco/nuon/pkg/helm"
	"github.com/nuonco/nuon/pkg/plans"
)

// Use the common diff package for the plan contents
type HelmPlanContents struct {
	Diff           string              `json:"plan"`
	Op             string              `json:"op"`
	ContentDiff    []diff.ResourceDiff `json:"helm_content_diff"`
	TemplateOutput string              `json:"template_output,omitempty"`
}

// Modify Exec function to use the common diff package
func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// Tag this handler's logger with semantic-convention attributes so every
	// emitted record (including those from helpers further down the call tree
	// that read the logger from ctx) carries them automatically.
	l = l.With(
		zap.String("service.name", "runner.helm"),
		zap.String("nuon.tool", "helm"),
		zap.String("nuon.deploy.kind", "helm"),
		zap.String("helm.release_name", h.state.plan.HelmDeployPlan.Name),
		zap.String("helm.namespace", h.state.plan.HelmDeployPlan.Namespace),
		zap.String("helm.chart_id", h.state.plan.HelmDeployPlan.HelmChartID),
	)
	ctx = pkgctx.SetLogger(ctx, l)

	l.Debug("Initializing Helm...",
		zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.state.arch.BasePath()},
	)
	actionCfg, kubeCfg, err := h.actionInit(ctx, l)
	if err != nil {
		return fmt.Errorf("unable to initialize helm actions: %w", err)
	}

	// set the release storage backend dynamically
	releaseStore, err := h.getHelmReleaseStore(ctx, kubeCfg)
	if err != nil {
		return errors.Wrap(err, "unable to get release store")
	}

	actionCfg.Releases = releaseStore

	l.Debug("Checking for previous Helm release...",
		zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.state.arch.BasePath()},
	)
	prevRel, err := helm.GetRelease(actionCfg, h.state.plan.HelmDeployPlan.Name)
	if err != nil {
		return fmt.Errorf("unable to get previous helm release: %w", err)
	}

	var (
		rel      *release.Release
		op       string
		diffStr  string
		helmPlan HelmPlanContents
	)

	// Load helm plan from the plan
	if len(h.state.plan.ApplyPlanContents) > 0 {
		// Use the new plans utility to decompress and decode the plan
		l.Debug("extracting apply plan contents", zap.Int("contents.compressed.length", len(h.state.plan.ApplyPlanContents)))
		decompressedPlan, err := plans.DecompressPlan(h.state.plan.ApplyPlanContents)
		if err != nil {
			return errors.Wrap(err, "unable to decompress apply plan contents")
		}

		if err := json.Unmarshal(decompressedPlan, &helmPlan); err != nil {
			return errors.Wrap(err, "unable to unmarshal apply plan contents")
		}

		l.Debug("extracting apply plan contents", zap.String("plan.op", helmPlan.Op))
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		var contentDiff *[]diff.ResourceDiff
		var templateOutput string
		var err error
		// in this case, the diff is generated so it is available to the createAPIResult method
		if prevRel == nil {
			helmPlan.Op = "install"
			l = l.With(zap.String("helm.operation", helmPlan.Op))
			opCtx, end := pkgop.Tool(ctx, "helm", "install_diff")
			opLog := pkgctx.LoggerOrDefault(opCtx, l)
			diffStr, contentDiff, templateOutput, err = h.installDiff(opCtx, opLog, actionCfg, kubeCfg)
			end(err)
		} else {
			helmPlan.Op = "upgrade"
			l = l.With(zap.String("helm.operation", helmPlan.Op))
			opCtx, end := pkgop.Tool(ctx, "helm", "upgrade_diff")
			opLog := pkgctx.LoggerOrDefault(opCtx, l)
			diffStr, contentDiff, templateOutput, err = h.upgrade_diff(opCtx, opLog, actionCfg, kubeCfg)
			end(err)
		}
		if err != nil {
			return err
		}

		if diffStr == "" {
			diffStr = "no changes"
		}

		helmPlan.Diff = diffStr
		helmPlan.ContentDiff = *contentDiff
		helmPlan.TemplateOutput = templateOutput

		l.Debug("calculated helm diff", zap.String("diff", diffStr))
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		// TODO(fd): figure out the best way to get a plan for this
		helmPlan.Op = "uninstall"
		l = l.With(zap.String("helm.operation", helmPlan.Op))
		l.Info("executing helm uninstall plan")

		opCtx, end := pkgop.Tool(ctx, "helm", "uninstall_diff")
		opLog := pkgctx.LoggerOrDefault(opCtx, l)
		diffStr, contentDiff, templateOutput, err := h.uninstallDiff(opCtx, opLog, actionCfg, kubeCfg, prevRel)
		end(err)
		if err != nil {
			return err
		}

		helmPlan.Diff = diffStr
		helmPlan.ContentDiff = *contentDiff
		helmPlan.TemplateOutput = templateOutput
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		l = l.With(zap.String("helm.operation", helmPlan.Op))
		l.Info(fmt.Sprintf("executing helm %s", helmPlan.Op))
		switch helmPlan.Op {
		case "install":
			if helm.ShouldUpgrade(prevRel) {
				l.Info("plan says install but release exists, switching to upgrade",
					zap.String("status", string(prevRel.Info.Status)),
					zap.Int("version", prevRel.Version),
				)
				op = "upgrade"
				opCtx, end := pkgop.Tool(ctx, "helm", "upgrade")
				opLog := pkgctx.LoggerOrDefault(opCtx, l)
				rel, err = h.upgrade(opCtx, opLog, actionCfg, kubeCfg)
				end(err)
			} else {
				op = "install"
				opCtx, end := pkgop.Tool(ctx, "helm", "install")
				opLog := pkgctx.LoggerOrDefault(opCtx, l)
				rel, err = h.install(opCtx, opLog, actionCfg, kubeCfg)
				end(err)
			}
		case "upgrade":
			op = "upgrade"
			opCtx, end := pkgop.Tool(ctx, "helm", "upgrade")
			opLog := pkgctx.LoggerOrDefault(opCtx, l)
			rel, err = h.upgrade(opCtx, opLog, actionCfg, kubeCfg)
			end(err)
		case "uninstall":
			op = "uninstall"
			opCtx, end := pkgop.Tool(ctx, "helm", "uninstall")
			opLog := pkgctx.LoggerOrDefault(opCtx, l)
			err = h.execUninstall(opCtx, opLog, actionCfg, job, jobExecution)
			end(err)
		default:
			l.Error("plan did not define an Op. this is unexpected.")
		}
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}

	// handle error
	if err != nil {
		h.writeErrorResult(ctx, op, err)
		return fmt.Errorf("unable to %s helm chart: %w", op, err)
	}

	var apiRes *models.ServiceCreateRunnerJobExecutionResultRequest
	var planContents HelmPlanContents

	// save plan if its not apply job operation is not apply
	if job.Operation != models.AppRunnerJobOperationTypeApplyDashPlan {
		planContents = helmPlan
	}

	apiRes, err = h.createAPIResultRequest(l, rel, planContents)
	if err != nil {
		h.writeErrorResult(ctx, op, err)
		return fmt.Errorf("unable to create api result from release: %w", err)
	}

	_, err = h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, apiRes)
	if err != nil {
		l.Error("failed to create job executione result", zap.Error(err))
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

func (h *handler) execUninstall(
	ctx context.Context,
	l *zap.Logger,
	actionCfg *action.Configuration,
	job *models.AppRunnerJob,
	jobExecution *models.AppRunnerJobExecution,
) error {
	if err := h.uninstall(ctx, l, actionCfg); err != nil {
		h.writeErrorResult(ctx, "uninstall", err)
		return fmt.Errorf("unable to uninstall helm chart: %w", err)
	}

	res := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}
	if _, err := h.apiClient.CreateJobExecutionResult(
		ctx,
		job.ID,
		jobExecution.ID,
		res,
	); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

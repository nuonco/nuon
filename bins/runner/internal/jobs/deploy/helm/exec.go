package helm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/diff"
	"github.com/nuonco/nuon/pkg/helm"
	"github.com/nuonco/nuon/pkg/plans"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	pkgop "github.com/nuonco/nuon/pkg/runner/op"
)

func renderedManifest(rel *release.Release, plan HelmPlanContents) string {
	if rel != nil {
		return rel.Manifest
	}
	if plan.Op == "uninstall" {
		return ""
	}
	return plan.TemplateOutput
}

type HelmPlanContents struct {
	Diff           string              `json:"plan"`
	Op             string              `json:"op"`
	ContentDiff    []diff.ResourceDiff `json:"helm_content_diff"`
	TemplateOutput string              `json:"template_output,omitempty"`

	ReleaseStatus string `json:"helm_release_status,omitempty"`
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l = l.With(
		zap.String("service.name", "runner.helm"),
		zap.String("nuon.tool", "helm"),
		zap.String("nuon.deploy.kind", "helm"),
		zap.String("helm.release_name", h.state.plan.HelmDeployPlan.Name),
		zap.String("helm.namespace", h.state.plan.HelmDeployPlan.Namespace),
		zap.String("helm.chart_id", h.state.plan.HelmDeployPlan.HelmChartID),
	)
	ctx = pkgctx.SetLogger(ctx, l)

	if h.clusterProvider != nil {
		h.clusterProvider.Set(h.state.plan.HelmDeployPlan.ClusterInfo)
	}

	l.Debug("Initializing Helm...",
		zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.basePath()},
	)
	actionCfg, kubeCfg, err := h.actionInit(ctx, l)
	if err != nil {
		return fmt.Errorf("unable to initialize helm actions: %w", err)
	}

	releaseStore, err := h.getHelmReleaseStore(ctx, kubeCfg)
	if err != nil {
		return errors.Wrap(err, "unable to get release store")
	}

	actionCfg.Releases = releaseStore

	if job.Operation == models.AppRunnerJobOperationTypeApplyDashPlan && h.isRecovery() {
		opCtx, end := pkgop.Tool(ctx, "helm", "recover")
		opLog := pkgctx.LoggerOrDefault(opCtx, l)
		res, err := h.recoverRelease(opCtx, opLog, actionCfg)
		end(err)
		if err != nil {
			h.writeErrorResult(ctx, "recover", err)
			return fmt.Errorf("unable to recover helm release: %w", err)
		}

		return h.writeRecoverResult(ctx, l, job, jobExecution, res)
	}

	l.Debug("Checking for previous Helm release...",
		zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.basePath()},
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

	if len(h.state.plan.ApplyPlanContents) > 0 {
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

	// why: A pending release breaks the plan's dry-run as well as the apply, and helm's
	// own errors there are ambiguous ("cannot reuse a name that is still in use"
	// also means a genuine name collision). Failing here instead names the status
	// and revision, which is what the operator needs to act on.
	if job.Operation == models.AppRunnerJobOperationTypeCreateDashApplyDashPlan && helm.IsPending(prevRel) {
		err := newPendingReleaseError(h.state.plan.HelmDeployPlan.Name, prevRel)
		h.writeErrorResult(ctx, "create-apply-plan", err)
		return err
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		var contentDiff *[]diff.ResourceDiff
		var templateOutput string
		var err error
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

		helmPlan.ContentDiff = *contentDiff
		helmPlan.TemplateOutput = templateOutput
		if prevRel != nil && prevRel.Info != nil {
			helmPlan.ReleaseStatus = string(prevRel.Info.Status)
			if len(*contentDiff) == 0 && prevRel.Info.Status != release.StatusDeployed {
				diffStr = fmt.Sprintf(
					"no manifest changes, but the release is %s and was never rolled out — re-applying",
					prevRel.Info.Status,
				)
			}
		}
		helmPlan.Diff = diffStr

		l.Debug("calculated helm diff", zap.String("diff", diffStr),
			zap.String("helm.release_status", helmPlan.ReleaseStatus))
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		// TODO(fd): figure out the best way to get a plan for this
		helmPlan.Op = "uninstall"
		l = l.With(zap.String("helm.operation", helmPlan.Op))
		l.Info("calculating helm uninstall plan with a dry run; nothing is removed until this plan is applied")

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

		if helmPlan.Op != "uninstall" && helm.IsPending(prevRel) {
			err := newPendingReleaseError(h.state.plan.HelmDeployPlan.Name, prevRel)
			h.writeErrorResult(ctx, helmPlan.Op, err)
			return err
		}

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

	if err != nil {
		h.writeErrorResult(ctx, op, err)
		return fmt.Errorf("unable to %s helm chart: %w", op, err)
	}

	// why: Hand the rendered kinds to the health engine: a chart shipping only custom
	// resources is otherwise invisible, since nothing else knows to list them.
	// The release secret is the only other copy and health is denied secret
	// reads, so the deploy is where this has to happen.
	//
	// Plan-only runs count too: a drift check renders the chart without applying
	// it, which is how a component picks up its kinds without being redeployed.
	if h.manifestKinds != nil {
		if manifest := renderedManifest(rel, helmPlan); manifest != "" {
			h.manifestKinds.Set(h.state.plan.ComponentID, manifest)
		}
	}

	var apiRes *models.ServiceCreateRunnerJobExecutionResultRequest
	var planContents HelmPlanContents

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

package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"

	"github.com/nuonco/nuon/pkg/helm"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	recoverOp = "recover"

	recoverActionNone      = "none"
	recoverActionRollback  = "rollback"
	recoverActionUninstall = "uninstall"
)

// The "is stuck in pending-" phrasing is the errparse signal for
// helm.pending_operation; changing it must change the parser too.
func newPendingReleaseError(releaseName string, rel *release.Release) error {
	return fmt.Errorf(
		"release %s is stuck in %s at revision %d from an operation that never finished; "+
			"helm refuses every further operation on it until it is recovered",
		releaseName, rel.Info.Status, rel.Version,
	)
}

// Readable before the apply contents are decoded, since it gates the chart fetch.
func (h *handler) isRecovery() bool {
	return h.state != nil &&
		h.state.plan != nil &&
		h.state.plan.HelmDeployPlan != nil &&
		h.state.plan.HelmDeployPlan.RecoverRelease
}

// Empty for a recovery, which fetches no archive; reading the nil archive panics.
func (h *handler) basePath() string {
	if h.state == nil || h.state.arch == nil {
		return ""
	}
	return h.state.arch.BasePath()
}

// recoverResult is what the recovery did, for the job result and the log stream.
type recoverResult struct {
	Action   string
	Revision int
	// After is empty when the release no longer exists.
	Before string
	After  string
}

// Summary renders the result for the job's plan contents and the log stream.
func (r *recoverResult) Summary(releaseName string) string {
	switch r.Action {
	case recoverActionRollback:
		return fmt.Sprintf("rolled release %s back from %s to revision %d", releaseName, r.Before, r.Revision)
	case recoverActionUninstall:
		return fmt.Sprintf("uninstalled release %s, which was stuck in %s with no revision to roll back to", releaseName, r.Before)
	default:
		if r.Before == "" {
			return fmt.Sprintf("no release named %s is stored, nothing to recover", releaseName)
		}
		return fmt.Sprintf("release %s is %s, not stuck mid-operation, so nothing was changed", releaseName, r.Before)
	}
}

// recoverRelease rolls a stuck release back to the last revision that finished a
// rollout, or removes it when none did. Refusing to touch a non-pending release
// is what makes it safe to expose and idempotent.
func (h *handler) recoverRelease(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) (*recoverResult, error) {
	releaseName := h.state.plan.HelmDeployPlan.Name

	rel, err := helm.GetRelease(actionCfg, releaseName)
	// Error before nil: a store failure read as "not installed" would report
	// success while the release stayed stuck.
	if err != nil {
		return nil, fmt.Errorf("unable to read release %s: %w", releaseName, err)
	}
	if rel == nil {
		l.Info("no release stored, nothing to recover", zap.String("release", releaseName))
		return &recoverResult{Action: recoverActionNone}, nil
	}

	before := string(rel.Info.Status)
	if !helm.IsPending(rel) {
		l.Info("release is not stuck mid-operation, leaving it alone",
			zap.String("release", releaseName),
			zap.String("status", before),
			zap.Int("revision", rel.Version),
		)
		return &recoverResult{Action: recoverActionNone, Before: before, After: before}, nil
	}

	l.Info("release is stuck mid-operation",
		zap.String("release", releaseName),
		zap.String("status", before),
		zap.Int("revision", rel.Version),
	)

	history, err := helm.History(actionCfg, releaseName)
	if err != nil {
		return nil, fmt.Errorf("unable to read history for release %s: %w", releaseName, err)
	}

	res := &recoverResult{Before: before}
	if revision, ok := helm.LastGoodRevision(history); ok {
		l.Info("rolling back to the last revision that finished a rollout",
			zap.String("release", releaseName),
			zap.Int("revision", revision),
		)
		if err := helm.Rollback(actionCfg, releaseName, revision, h.state.timeout); err != nil {
			return nil, err
		}
		res.Action = recoverActionRollback
		res.Revision = revision
	} else {
		l.Info("no revision finished a rollout, so removing the stuck release instead",
			zap.String("release", releaseName),
		)
		if err := h.uninstall(ctx, l, actionCfg); err != nil {
			return nil, fmt.Errorf("unable to remove stuck release %s: %w", releaseName, err)
		}
		res.Action = recoverActionUninstall
	}

	res.After = h.releaseStatusAfterRecovery(l, actionCfg, releaseName)
	l.Info(res.Summary(releaseName), zap.String("release.status", res.After))

	return res, nil
}

// The release status rides in the plan contents, the existing channel for it.
func (h *handler) writeRecoverResult(
	ctx context.Context,
	l *zap.Logger,
	job *models.AppRunnerJob,
	jobExecution *models.AppRunnerJobExecution,
	res *recoverResult,
) error {
	contents := HelmPlanContents{
		Op:            recoverOp,
		Diff:          res.Summary(h.state.plan.HelmDeployPlan.Name),
		ReleaseStatus: res.After,
	}

	apiRes, err := h.createAPIResultRequest(l, nil, contents)
	if err != nil {
		h.writeErrorResult(ctx, recoverOp, err)
		return fmt.Errorf("unable to create api result for recovery: %w", err)
	}

	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, apiRes); err != nil {
		l.Error("failed to create job execution result", zap.Error(err))
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

// A read failure is not fatal: the recovery already succeeded.
func (h *handler) releaseStatusAfterRecovery(l *zap.Logger, actionCfg *action.Configuration, releaseName string) string {
	rel, err := helm.GetRelease(actionCfg, releaseName)
	if err != nil {
		l.Warn("unable to re-read release after recovery", zap.Error(err))
		return ""
	}
	if rel == nil || rel.Info == nil {
		return ""
	}

	return string(rel.Info.Status)
}

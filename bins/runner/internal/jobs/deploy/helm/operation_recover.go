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
	// recoverOp is the plan-contents op name for a recovery, alongside helm's
	// own install/upgrade/uninstall values.
	recoverOp = "recover"

	recoverActionNone      = "none"
	recoverActionRollback  = "rollback"
	recoverActionUninstall = "uninstall"
)

// newPendingReleaseError describes a release helm left mid-operation. The
// wording is the contract the control plane's error parser keys on to classify
// the failure as helm.pending_operation, so the "is stuck in pending-" phrasing
// must not drift without updating the parser's signal alongside it.
func newPendingReleaseError(releaseName string, rel *release.Release) error {
	return fmt.Errorf(
		"release %s is stuck in %s at revision %d from an operation that never finished; "+
			"helm refuses every further operation on it until it is recovered",
		releaseName, rel.Info.Status, rel.Version,
	)
}

// isRecovery reports whether this job is a recovery rather than a deploy. It
// gates the chart fetch and unpack as well as the operation itself, so it has to
// be readable before the plan's apply contents are decoded.
func (h *handler) isRecovery() bool {
	return h.state != nil &&
		h.state.plan != nil &&
		h.state.plan.HelmDeployPlan != nil &&
		h.state.plan.HelmDeployPlan.Recover != nil
}

// basePath is the unpacked chart's directory, or "" for a recovery, which never
// fetches an archive. It exists so logging the path cannot depend on whether an
// archive was fetched — reading it off the nil archive directly would panic on
// every recovery.
func (h *handler) basePath() string {
	if h.state == nil || h.state.arch == nil {
		return ""
	}
	return h.state.arch.BasePath()
}

// recoverResult is what the recovery did, so the job result and the log stream
// can both say it plainly rather than leaving the operator to infer it.
type recoverResult struct {
	Action   string
	Revision int
	// Before and After are the release statuses either side of the operation.
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

// recoverRelease returns a release that helm parked mid-operation to a usable
// state. It rolls back to the last revision that finished a rollout, or removes
// the release when there is no such revision — a first install that never rolled
// out has nothing behind it to return to.
//
// It refuses to touch a release that is not pending. That is the property that
// makes this safe to expose as a break-glass button: it cannot revert a healthy
// release, and running it twice is a no-op the second time.
func (h *handler) recoverRelease(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) (*recoverResult, error) {
	releaseName := h.state.plan.HelmDeployPlan.Name

	rel, err := helm.GetRelease(actionCfg, releaseName)
	// The error is checked before the nil release: a store or API failure that
	// was read as "not installed" would make a recovery silently report success
	// while the release stayed stuck.
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

// writeRecoverResult reports the recovery to the control plane. The resulting
// release status rides along in the plan contents, which is the existing channel
// for it, so the dashboard can drop its stuck-release banner without a new field
// on the job result.
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

// releaseStatusAfterRecovery re-reads the release so the control plane records
// where recovery actually left it. A read failure here is not fatal: the
// recovery already succeeded, and reporting an unknown status is better than
// failing a job that did its work.
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

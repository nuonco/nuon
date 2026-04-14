package helm

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	composite_errors "github.com/nuonco/nuon/bins/runner/internal/pkg/composite_errors"
	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) writeErrorResult(ctx context.Context, step string, err error) {
	resultReq := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:   false,
		ErrorCode: 0,
		ErrorMetadata: map[string]string{
			"step":     step,
			"handler":  h.Name(),
			"job_type": string(h.JobType()),
			"message":  err.Error(),
		},
	}

	if _, err := h.apiClient.CreateJobExecutionResult(ctx, h.state.jobID, h.state.jobExecutionID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}
}

func (h *handler) reportCompositeErrors(ctx context.Context, ownerType string, helmErr error) {
	errs := composite_errors.ParseHelmStderr(helmErr.Error(), ownerType)
	if len(errs) == 0 {
		errs = composite_errors.FromGoError(helmErr, ownerType)
	}

	// For deploy failures, also collect K8s diagnostics (pod statuses, events, logs)
	if ownerType == "apply" {
		k8sErrs := h.collectK8sDiagnostics(ctx)
		errs = append(errs, k8sErrs...)
	}

	modelErrs := composite_errors.ToModels(errs)
	if err := h.apiClient.ReportCompositeErrors(ctx, h.state.jobID, modelErrs); err != nil {
		h.errRecorder.Record("report composite errors", err)
	}
}

func (h *handler) collectK8sDiagnostics(ctx context.Context) []composite_errors.CompositeError {
	l, _ := pkgctx.Logger(ctx)
	if l == nil {
		l = zap.NewNop()
	}

	if h.state.plan == nil || h.state.plan.HelmDeployPlan == nil {
		return nil
	}

	clusterInfo := h.state.plan.HelmDeployPlan.ClusterInfo
	if clusterInfo == nil {
		return nil
	}

	kubeCfg, err := kube.ConfigForCluster(ctx, clusterInfo)
	if err != nil {
		l.Warn("unable to create kube config for k8s diagnostics", zap.Error(err))
		return nil
	}

	k8sClient, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		l.Warn("unable to create k8s client for diagnostics", zap.Error(err))
		return nil
	}

	namespace := h.state.plan.HelmDeployPlan.Namespace
	releaseName := h.state.plan.HelmDeployPlan.Name

	l.Info("collecting k8s diagnostics after helm failure",
		zap.String("namespace", namespace),
		zap.String("release", releaseName))

	return composite_errors.CollectHelmK8sDiagnostics(ctx, k8sClient, namespace, releaseName)
}

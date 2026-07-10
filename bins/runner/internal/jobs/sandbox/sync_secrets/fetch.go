package terraform

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	h.state = &handlerState{}

	l.Info("fetching sync secrets job plan")
	cp, err := h.apiClient.GetJobCompositePlan(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	composite, err := plantypes.CompositePlanFromAny(cp)
	if err != nil {
		return errors.Wrap(err, "unable to parse composite plan")
	}
	if composite.SyncSecretsPlan == nil {
		return errors.New("composite plan missing sync secrets plan")
	}
	plan := composite.SyncSecretsPlan
	h.state.plan = plan

	if h.state.plan.ClusterInfo != nil {
		h.state.plan.ClusterInfo.WithAWSAuth(h.state.plan.AWSAuth)
		h.state.plan.ClusterInfo.WithAzureAuth(h.state.plan.AzureAuth)
		h.state.plan.ClusterInfo.WithGCPAuth(h.state.plan.GCPAuth)
	}

	l.Info("setting sandbox operation timeout", zap.String("duration", h.state.timeout.String()))
	return nil
}

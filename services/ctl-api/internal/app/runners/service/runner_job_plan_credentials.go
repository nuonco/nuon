package service

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/registryauth"
)

func (s *service) refreshDemandPlanCredentials(ctx context.Context, job *app.RunnerJob, cp *plantypes.CompositePlan) ([]registryauth.Refreshed, error) {
	if demand.SignalID(job) == "" || cp == nil {
		return nil, nil
	}

	l := s.l.With(zap.String("runner_job_id", job.ID), zap.String("org_id", job.OrgID))

	var org app.Org
	if err := s.db.WithContext(ctx).Where(app.Org{ID: job.OrgID}).First(&org).Error; err != nil {
		return nil, err
	}
	if org.SandboxMode {
		return nil, nil
	}

	refreshed, err := s.registryAuth.RefreshCompositePlan(ctx, l, job.OrgID, cp)
	if err != nil {
		return nil, err
	}
	urls := make(map[string]string)
	refreshGit := func(source *plantypes.GitSource, path ...string) error {
		if source == nil {
			return nil
		}
		oldURL := source.URL
		if fresh, ok := urls[oldURL]; ok {
			source.URL = fresh
		} else {
			if err := s.vcsHelpers.RefreshGitSource(ctx, job.OrgID, source); err != nil {
				return err
			}
			urls[oldURL] = source.URL
		}
		if source.URL != oldURL {
			refreshed = append(refreshed, registryauth.Refreshed{Path: path, Value: source.URL})
		}
		return nil
	}
	if cp.BuildPlan != nil {
		if err := refreshGit(cp.BuildPlan.Src, "build_plan", "git_source", "url"); err != nil {
			return nil, err
		}
	}
	if cp.SandboxRunPlan != nil {
		if err := refreshGit(cp.SandboxRunPlan.GitSource, "sandbox_run_plan", "git_source", "url"); err != nil {
			return nil, err
		}
	}
	if cp.ActionWorkflowRunPlan != nil {
		for idx, step := range cp.ActionWorkflowRunPlan.Steps {
			if err := refreshGit(step.GitSource, "action_workflow_run_plan", "steps", strconv.Itoa(idx), "git_source", "url"); err != nil {
				return nil, err
			}
		}
	}
	return refreshed, nil
}

func (s *service) refreshDemandPlanJSON(ctx context.Context, runnerPlan *app.RunnerJobPlan, job *app.RunnerJob) (string, error) {
	if demand.SignalID(job) == "" || runnerPlan.PlanJSON == "" {
		return runnerPlan.PlanJSON, nil
	}

	cp, err := runnerPlan.DeriveCompositePlan(job)
	if err != nil {
		return "", err
	}

	refreshed, err := s.refreshDemandPlanCredentials(ctx, job, cp)
	if err != nil {
		return "", err
	}
	if len(refreshed) == 0 {
		return runnerPlan.PlanJSON, nil
	}

	patched, err := registryauth.PatchSubPlanJSON([]byte(runnerPlan.PlanJSON), refreshed)
	if err != nil {
		return "", err
	}

	return string(patched), nil
}

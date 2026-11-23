package ociprune

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/oci/prune"
)

type Outcome string

const (
	OutcomeDryRun           Outcome = "dry_run"
	OutcomePruned           Outcome = "pruned"
	OutcomePartial          Outcome = "partial"
	OutcomePermissionDenied Outcome = "permission_denied"
)

type Result struct {
	Outcome    Outcome                  `json:"outcome"`
	DryRun     bool                     `json:"dry_run"`
	Message    string                   `json:"message,omitempty"`
	Manifests  int                      `json:"manifests"`
	Kept       int                      `json:"kept"`
	Candidates int                      `json:"candidates"`
	Truncated  bool                     `json:"truncated"`
	Planned    []prune.Deletion         `json:"planned,omitempty"`
	Deleted    []string                 `json:"deleted,omitempty"`
	Failures   []prune.DeleteFailure    `json:"failures,omitempty"`
	KeptByWhy  map[prune.KeepReason]int `json:"kept_by_reason,omitempty"`
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}
	plan := h.state.plan

	repo, err := prune.NewECRRepository(ctx, plan.Registry)
	if err != nil {
		h.writeErrorResult(ctx, "connect to registry", err)
		return err
	}

	manifests, err := repo.ListManifests(ctx)
	if errors.Is(err, prune.ErrPermissionDenied) {
		return h.finish(ctx, job, jobExecution, &Result{
			Outcome: OutcomePermissionDenied,
			DryRun:  plan.DryRun,
			Message: err.Error(),
		})
	}
	if err != nil {
		h.writeErrorResult(ctx, "list registry", err)
		return err
	}
	l.Info("listed registry", zap.String("repository", repo.URL), zap.Int("manifests", len(manifests)))

	inUse, err := prune.ScanCluster(ctx, plan.ClusterInfo, repo.URL)
	if err != nil {
		h.writeErrorResult(ctx, "scan cluster", err)
		return err
	}
	l.Info("scanned cluster", zap.Int("tags_in_use", len(inUse.Tags)), zap.Int("digests_in_use", len(inUse.Digests)))

	decision := prune.Decide(prune.Input{
		Manifests:    manifests,
		OwnedTags:    plan.OwnedTags,
		KeepTags:     plan.KeepTags,
		InUseTags:    inUse.Tags,
		InUseDigests: inUse.Digests,
		GracePeriod:  time.Duration(plan.GracePeriodSeconds) * time.Second,
		Now:          time.Now(),
		MaxDeletes:   plan.MaxDeletes,
	})

	res := &Result{
		Outcome:    OutcomeDryRun,
		DryRun:     plan.DryRun,
		Manifests:  len(manifests),
		Kept:       len(decision.Kept),
		Candidates: decision.Total,
		Truncated:  decision.Truncated,
		Planned:    decision.Deletions(),
		KeptByWhy:  map[prune.KeepReason]int{},
	}
	for _, reason := range decision.Kept {
		res.KeptByWhy[reason]++
	}
	for _, d := range res.Planned {
		l.Info("prune candidate", zap.String("digest", d.Digest), zap.Strings("tags", d.Tags), zap.Bool("dry_run", plan.DryRun))
	}

	if plan.DryRun {
		return h.finish(ctx, job, jobExecution, res)
	}

	res.Outcome = OutcomePruned
	for _, batch := range decision.Batches {
		deleted, failures, err := repo.Delete(ctx, batch)
		res.Deleted = append(res.Deleted, deleted...)
		res.Failures = append(res.Failures, failures...)
		if errors.Is(err, prune.ErrPermissionDenied) {
			res.Outcome = OutcomePermissionDenied
			res.Message = err.Error()
			return h.finish(ctx, job, jobExecution, res)
		}
		if err != nil {
			h.writeErrorResult(ctx, "delete images", err)
			return err
		}
		if len(failures) > 0 {
			res.Outcome = OutcomePartial
			break
		}
	}
	l.Info("pruned registry", zap.Int("deleted", len(res.Deleted)), zap.Int("failures", len(res.Failures)))

	return h.finish(ctx, job, jobExecution, res)
}

func (h *handler) finish(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution, res *Result) error {
	h.state.result = res

	contents, err := json.Marshal(res)
	if err != nil {
		return err
	}
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:  true,
		Contents: string(contents),
	}); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}
	return nil
}

package actionworkflowrun

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/imagesync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

func (s *Signal) runActionSequence(ctx workflow.Context, installID string, run *app.InstallActionWorkflowRun, logStreamID string, compositeErrors bool) error {
	deps, err := s.decideImageDeps(ctx, run)
	if err != nil {
		if compositeErrors {
			s.recordPreparationCompositeError(ctx, run.ID, err)
		}
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStreamID)
		return err
	}

	var phases []job.Phase
	for _, dep := range deps.Deps {
		phases = append(phases, s.imageDepPhase(run, logStreamID, dep))
	}
	phases = append(phases,
		job.Phase{
			Key: checkpointImageMirror,
			Prepare: func(ctx workflow.Context) (string, error) {
				if !deps.Published {
					if err := publishImageDeps(ctx, run, deps.Deps); err != nil {
						s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
						return "", err
					}
					deps.Published = true
					if err := job.SaveCheckpoint(ctx, checkpointImageDeps, deps); err != nil {
						return "", err
					}
				}
				return s.prepareMirrorPhase(ctx, run, logStreamID, compositeErrors)
			},
			Finalize: func(ctx workflow.Context, jobID string) error {
				_, err := s.finalizeMirrorJob(ctx, run, jobID)
				if err != nil {
					s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
				}
				return err
			},
		},
		job.Phase{
			Key: checkpointAction,
			Prepare: func(ctx workflow.Context) (string, error) {
				return s.prepareActionPhase(ctx, installID, run, logStreamID, compositeErrors)
			},
			Finalize: func(ctx workflow.Context, _ string) error {
				return s.finalizeActionRun(ctx, installID, run.ID)
			},
		},
	)

	return (job.Sequence{
		Phases: phases,
		Finish: func(ctx workflow.Context, err error) error {
			if job.IsRunnerJobFailure(err) {
				for _, dep := range deps.Deps {
					var cp job.Checkpoint
					if _, loadErr := job.LoadCheckpoint(ctx, dep.phaseKey(), &cp); loadErr == nil && !cp.Finalized {
						var prep imagesync.SyncPreparation
						if _, loadErr := job.LoadCheckpoint(ctx, dep.phaseKey()+"-preparation", &prep); loadErr == nil {
							imagesync.FailSync(ctx, prep)
						}
						break
					}
				}
				s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, job.JobErrorMessage(err, "action workflow failed"))
			}
			activities.AwaitCloseLogStreamByLogStreamID(ctx, logStreamID)
			return err
		},
	}).Run(ctx)
}

func (s *Signal) decideImageDeps(ctx workflow.Context, run *app.InstallActionWorkflowRun) (imageDepsCheckpoint, error) {
	var cp imageDepsCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointImageDeps, &cp); err != nil || cp.Decided {
		return cp, err
	}
	in, err := s.imageDepInputs(ctx, run)
	if err != nil {
		return cp, err
	}
	if in != nil {
		cp.Deps, err = decideImageDeps(ctx, in)
		if err != nil {
			return cp, err
		}
	}
	cp.Decided = true
	cp.Published = len(cp.Deps) == 0
	return cp, job.SaveCheckpoint(ctx, checkpointImageDeps, cp)
}

func (s *Signal) imageDepPhase(run *app.InstallActionWorkflowRun, logStreamID string, dep imageDepCheckpoint) job.Phase {
	return job.Phase{
		Key: dep.phaseKey(),
		Prepare: func(ctx workflow.Context) (string, error) {
			prep, jobID, err := imagesync.PrepareSync(ctx, imagesync.SyncRequest{
				Install: &run.Install, ComponentID: dep.ComponentID, BuildID: dep.BuildID,
				FlowID: s.InstallWorkflowID, ParentLogStreamID: logStreamID,
				WorkflowIDSuffix: "-action-image-dep-" + run.ID + "-" + dep.ComponentID,
			})
			if err == nil {
				err = job.SaveCheckpoint(ctx, dep.phaseKey()+"-preparation", prep)
			}
			if err != nil {
				s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
			}
			return jobID, err
		},
		Finalize: func(ctx workflow.Context, jobID string) error {
			var prep imagesync.SyncPreparation
			if _, err := job.LoadCheckpoint(ctx, dep.phaseKey()+"-preparation", &prep); err != nil {
				return err
			}
			err := imagesync.FinalizeSync(ctx, prep, jobID)
			if err != nil {
				s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
			}
			return err
		},
	}
}

func (s *Signal) pinActionPlan(ctx workflow.Context, run *app.InstallActionWorkflowRun, compositeErrors bool) (*plan.CreateActionPlanResponse, error) {
	var pinned plan.CreateActionPlanResponse
	if ok, err := job.LoadCheckpoint(ctx, checkpointActionPlan, &pinned); err != nil || ok {
		return &pinned, err
	}
	created, err := s.createActionPlan(ctx, run, compositeErrors)
	if err != nil {
		return nil, err
	}
	return created, job.SaveCheckpoint(ctx, checkpointActionPlan, created)
}

func (s *Signal) prepareMirrorPhase(ctx workflow.Context, run *app.InstallActionWorkflowRun, logStreamID string, compositeErrors bool) (string, error) {
	response, err := s.pinActionPlan(ctx, run, compositeErrors)
	if err != nil {
		return "", err
	}
	if response.Plan.SourceImage == "" {
		return "", nil
	}
	if err := s.checkImageActionSupported(ctx, run); err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
		return "", err
	}
	if response.Plan.ImageDigestRef != "" {
		return "", nil
	}
	jobID, err := s.prepareMirrorJob(ctx, run, logStreamID, response.Plan)
	if err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
	}
	return jobID, err
}

func (s *Signal) prepareActionPhase(ctx workflow.Context, installID string, run *app.InstallActionWorkflowRun, logStreamID string, compositeErrors bool) (string, error) {
	response, err := s.pinActionPlan(ctx, run, compositeErrors)
	if err != nil {
		return "", err
	}
	var mirror job.Checkpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointImageMirror, &mirror); err != nil {
		return "", err
	}
	if mirror.JobID != "" {
		response.Plan.ImageDigestRef, _, err = resolveMirroredDigestRef(ctx, mirror.JobID)
		if err != nil {
			s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
			return "", errors.Wrap(err, "unable to resolve action image")
		}
	}
	return s.createActionJob(ctx, installID, run, logStreamID, response, compositeErrors)
}

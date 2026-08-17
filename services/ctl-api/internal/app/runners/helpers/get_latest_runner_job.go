package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type GetLatestJobRequest struct {
	OwnerID string
}

func (a *Helpers) GetLatestJob(ctx context.Context, req *GetLatestJobRequest) (*app.RunnerJob, error) {
	var job app.RunnerJob
	res := a.db.WithContext(ctx).
		Where(app.RunnerJob{
			OwnerID: req.OwnerID,
		}).
		Preload("Executions").
		Preload("Executions.Result", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_execution_results.created_at DESC")
		}).
		Order("created_at desc").
		Limit(1).
		First(&job)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return &job, nil
}

type GetLatestJobCompositeErrorRequest struct {
	OwnerID   string
	OwnerType string
}

func GetLatestJobCompositeError(ctx context.Context, db *gorm.DB, req GetLatestJobCompositeErrorRequest) (*compositeerrors.CompositeErrorData, error) {
	var job app.RunnerJob
	res := db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Where(app.RunnerJob{
			OwnerID:   req.OwnerID,
			OwnerType: req.OwnerType,
		}).
		Preload("Executions", func(db *gorm.DB) *gorm.DB {
			return db.
				Select("id", "runner_job_id", "created_at").
				Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
				Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
				Limit(1)
		}).
		Preload("Executions.Result", func(db *gorm.DB) *gorm.DB {
			return db.Select("runner_job_execution_id", "composite_error")
		}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
		First(&job)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get latest runner job composite error: %w", res.Error)
	}

	return ResolveJobCompositeError(&job), nil
}

func ResolveJobCompositeError(job *app.RunnerJob) *compositeerrors.CompositeErrorData {
	if job.Status == app.RunnerJobStatusCancelled {
		if job.CompositeError != nil && job.CompositeError.Type == joberrors.CancellationErrorType {
			return job.CompositeError
		}
		return nil
	}

	switch job.Status {
	case app.RunnerJobStatusFailed, app.RunnerJobStatusTimedOut, app.RunnerJobStatusNotAttempted:
		if len(job.Executions) > 0 && job.Executions[0].Result != nil && job.Executions[0].Result.CompositeError != nil {
			return job.Executions[0].Result.CompositeError
		}
		return job.CompositeError
	default:
		return nil
	}
}

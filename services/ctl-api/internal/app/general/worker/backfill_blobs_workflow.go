package worker

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/workflows/blobbackfill"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	blobBackfillBatchSize = 1000
	// caps batches per child and buckets per parent run so workflow history stays bounded.
	blobBackfillMaxBatches = 5000
	blobBucketsPerRun      = 100
)

var defaultBlobBackfillTables = []string{
	"install_states",
	"install_workflow_step_approvals",
	"runner_job_plans",
	"runner_job_execution_outputs",
	"terraform_workspace_states",
	"terraform_workspace_state_jsons",
}

// BackfillBlobs enumerates the day-buckets with un-mirrored rows and drains them
// one (table, day) child at a time, continue-as-newing to keep history bounded.
// Running sequentially keeps the per-activity S3 rate limiter acting globally.
func (w *Workflows) BackfillBlobs(ctx workflow.Context, req blobbackfill.RangeRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	if req.Processed == nil {
		req.Processed = map[string]int64{}
	}

	progress := blobbackfill.Progress{
		Processed: req.Processed,
		DaysTotal: req.DaysTotal,
		DaysDone:  req.DaysDone,
	}
	if err := workflow.SetQueryHandler(ctx, blobbackfill.ProgressQueryType, func() (blobbackfill.Progress, error) {
		return progress, nil
	}); err != nil {
		return errors.Wrap(err, "unable to register blob backfill progress query handler")
	}

	if !req.Initialized {
		l.Info("general workflow execution", zap.String("type", "blob-backfill"))

		tables := req.Tables
		if len(tables) == 0 {
			tables = defaultBlobBackfillTables
		}

		for _, table := range tables {
			if _, ok := req.Processed[table]; !ok {
				req.Processed[table] = 0
			}

			resp, err := activities.AwaitListBackfillDays(ctx, activities.ListBlobDaysRequest{Table: table})
			if err != nil {
				return errors.Wrapf(err, "unable to list backfill days for %s", table)
			}
			for _, day := range resp.Days {
				req.Pending = append(req.Pending, blobbackfill.DayBucket{Table: table, Day: day})
			}
		}

		req.Tables = tables
		req.Initialized = true
		req.DaysTotal = len(req.Pending)
		progress.DaysTotal = req.DaysTotal
	}

	processedThisRun := 0
	for len(req.Pending) > 0 && processedThisRun < blobBucketsPerRun {
		bucket := req.Pending[0]
		req.Pending = req.Pending[1:]

		progress.CurrentTable = bucket.Table
		progress.CurrentDay = bucket.Day

		result, err := w.runBackfillDay(ctx, bucket)
		if err != nil {
			return errors.Wrapf(err, "unable to backfill %s day %s", bucket.Table, bucket.Day)
		}

		req.Processed[bucket.Table] += result.Processed
		req.DaysDone++
		progress.DaysDone = req.DaysDone
		processedThisRun++
	}

	if len(req.Pending) > 0 {
		l.Info("continuing blob backfill", zap.Int("days_done", req.DaysDone), zap.Int("days_total", req.DaysTotal))
		return workflow.NewContinueAsNewError(ctx, blobbackfill.WorkflowName, req)
	}

	progress.CurrentTable = ""
	progress.CurrentDay = ""
	progress.Done = true
	l.Info("backfilled blobs", zap.Int("days", req.DaysDone))
	return nil
}

func (w *Workflows) runBackfillDay(ctx workflow.Context, bucket blobbackfill.DayBucket) (blobbackfill.DayResult, error) {
	childID := fmt.Sprintf("%s-%s-%s", blobbackfill.WorkflowID, bucket.Table, bucket.Day)
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:            childID,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	})

	var result blobbackfill.DayResult
	if err := workflow.ExecuteChildWorkflow(childCtx, blobbackfill.DayWorkflowName, blobbackfill.DayRequest{
		Table: bucket.Table,
		Day:   bucket.Day,
	}).Get(childCtx, &result); err != nil {
		return result, err
	}
	return result, nil
}

// BackfillBlobsDay drains every un-mirrored row created within its day.
func (w *Workflows) BackfillBlobsDay(ctx workflow.Context, req blobbackfill.DayRequest) (blobbackfill.DayResult, error) {
	var result blobbackfill.DayResult
	for i := 0; i < blobBackfillMaxBatches; i++ {
		resp, err := activities.AwaitBackfillBlobs(ctx, activities.BackfillBlobsRequest{
			Table:     req.Table,
			Day:       req.Day,
			BatchSize: blobBackfillBatchSize,
		})
		if err != nil {
			return result, errors.Wrapf(err, "unable to backfill blobs for %s day %s", req.Table, req.Day)
		}

		result.Processed += resp.Processed
		if resp.Processed < blobBackfillBatchSize {
			break
		}
	}
	return result, nil
}

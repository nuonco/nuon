package worker

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/workflows/blobverify"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	blobVerifyBatchSize  = 1000
	blobVerifyMaxBatches = 5000
	// maxMismatchIDSamples caps retained mismatched ids so parent history stays small.
	maxMismatchIDSamples = 100
)

// VerifyBlobs enumerates the day-buckets with rows and walks them one
// (table, day) child at a time, continue-as-newing to keep history bounded.
// Running sequentially keeps the per-activity S3 rate limiter acting globally.
func (w *Workflows) VerifyBlobs(ctx workflow.Context, req blobverify.RangeRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	if req.Tallies == nil {
		req.Tallies = map[string]blobverify.TableProgress{}
	}

	progress := blobverify.Progress{
		Tables:    req.Tallies,
		DaysTotal: req.DaysTotal,
		DaysDone:  req.DaysDone,
	}
	if err := workflow.SetQueryHandler(ctx, blobverify.ProgressQueryType, func() (blobverify.Progress, error) {
		return progress, nil
	}); err != nil {
		return errors.Wrap(err, "unable to register blob verify progress query handler")
	}

	if !req.Initialized {
		l.Info("general workflow execution", zap.String("type", "blob-verify"))

		tables := req.Tables
		if len(tables) == 0 {
			tables = defaultBlobBackfillTables
		}

		for _, table := range tables {
			if _, ok := req.Tallies[table]; !ok {
				req.Tallies[table] = blobverify.TableProgress{}
			}

			resp, err := activities.AwaitListVerifyDays(ctx, activities.ListBlobDaysRequest{Table: table})
			if err != nil {
				return errors.Wrapf(err, "unable to list verify days for %s", table)
			}
			for _, day := range resp.Days {
				req.Pending = append(req.Pending, blobverify.DayBucket{Table: table, Day: day})
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

		result, err := w.runVerifyDay(ctx, bucket)
		if err != nil {
			return errors.Wrapf(err, "unable to verify %s day %s", bucket.Table, bucket.Day)
		}

		tp := req.Tallies[bucket.Table]
		tp.Checked += result.Checked
		tp.Mismatched += result.Mismatched
		tp.NotSet += result.NotSet
		for _, id := range result.MismatchedIDs {
			if len(tp.MismatchedIDs) < maxMismatchIDSamples {
				tp.MismatchedIDs = append(tp.MismatchedIDs, id)
			}
		}
		req.Tallies[bucket.Table] = tp

		req.DaysDone++
		progress.DaysDone = req.DaysDone
		processedThisRun++
	}

	if len(req.Pending) > 0 {
		l.Info("continuing blob verify", zap.Int("days_done", req.DaysDone), zap.Int("days_total", req.DaysTotal))
		return workflow.NewContinueAsNewError(ctx, blobverify.WorkflowName, req)
	}

	progress.CurrentTable = ""
	progress.CurrentDay = ""
	progress.Done = true
	l.Info("verified blobs", zap.Int("days", req.DaysDone))
	return nil
}

func (w *Workflows) runVerifyDay(ctx workflow.Context, bucket blobverify.DayBucket) (blobverify.DayResult, error) {
	childID := fmt.Sprintf("%s-%s-%s", blobverify.WorkflowID, bucket.Table, bucket.Day)
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:            childID,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	})

	var result blobverify.DayResult
	if err := workflow.ExecuteChildWorkflow(childCtx, blobverify.DayWorkflowName, blobverify.DayRequest{
		Table: bucket.Table,
		Day:   bucket.Day,
	}).Get(childCtx, &result); err != nil {
		return result, err
	}
	return result, nil
}

// VerifyBlobsDay verifies every row created within its day, walking by id cursor.
func (w *Workflows) VerifyBlobsDay(ctx workflow.Context, req blobverify.DayRequest) (blobverify.DayResult, error) {
	var result blobverify.DayResult
	cursor := ""
	for i := 0; i < blobVerifyMaxBatches; i++ {
		resp, err := activities.AwaitVerifyBlobs(ctx, activities.VerifyBlobsRequest{
			Table:     req.Table,
			Day:       req.Day,
			BatchSize: blobVerifyBatchSize,
			Cursor:    cursor,
		})
		if err != nil {
			return result, errors.Wrapf(err, "unable to verify blobs for %s day %s", req.Table, req.Day)
		}

		result.Checked += resp.Checked
		result.Mismatched += resp.Mismatched
		result.NotSet += resp.NotSet
		for _, m := range resp.Mismatches {
			if len(result.MismatchedIDs) < maxMismatchIDSamples {
				result.MismatchedIDs = append(result.MismatchedIDs, m.ID)
			}
		}

		if resp.Checked < blobVerifyBatchSize {
			break
		}
		cursor = resp.NextCursor
	}
	return result, nil
}

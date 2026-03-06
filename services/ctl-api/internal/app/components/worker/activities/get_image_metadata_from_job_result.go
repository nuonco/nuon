package activities

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/oci/metadata"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetImageMetadataFromJobResultRequest struct {
	JobID string `validate:"required"`
}

type GetImageMetadataFromJobResultResponse struct {
	Metadata *metadata.ImageMetadata `json:"metadata" temporaljson:"metadata,omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 2
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 30s
func (a *Activities) GetImageMetadataFromJobResult(ctx context.Context, req *GetImageMetadataFromJobResultRequest) (*GetImageMetadataFromJobResultResponse, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("job_id", req.JobID))

	l.Info("getting image metadata from job result")

	// Get the job with its executions and results
	var job app.RunnerJob
	res := a.db.WithContext(ctx).
		Preload("Executions", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_executions.created_at DESC").Limit(1)
		}).
		Preload("Executions.Result").
		First(&job, "id = ?", req.JobID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get job")
	}

	if len(job.Executions) == 0 {
		return nil, fmt.Errorf("job %s has no executions", req.JobID)
	}

	execution := job.Executions[0]
	if execution.Result == nil {
		return nil, fmt.Errorf("job execution %s has no result", execution.ID)
	}

	result := execution.Result

	// Decompress the gzipped contents
	if len(result.ContentsGzip) == 0 {
		return nil, fmt.Errorf("job execution result has no compressed contents")
	}

	reader, err := gzip.NewReader(bytes.NewReader(result.ContentsGzip))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create gzip reader")
	}
	defer reader.Close()

	metadataJSON, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decompress metadata")
	}

	var imgMetadata metadata.ImageMetadata
	if err := json.Unmarshal(metadataJSON, &imgMetadata); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal image metadata")
	}

	l.Info("image metadata retrieved successfully",
		zap.String("digest", imgMetadata.Digest),
		zap.Bool("signed", imgMetadata.Signed),
	)

	return &GetImageMetadataFromJobResultResponse{
		Metadata: &imgMetadata,
	}, nil
}

package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const PageSize int = 10000

type Page struct {
	Data  []app.OtelLogRecord `json:"data"`
	Next  string              `json:"next"`
	Count int64               `json:"count"`
}

// @ID OtelReadLogs
// @Summary	get a runner's logs
// @Description.markdown runner_otel_read_logs.md
// @Param			runner_id	path	string	true	"runner ID"
// @Param   job_id query string false	"job id"
// @Param   job_execution_id query string false	"job execution id"
// @Tags runners
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	[]app.OtelLogRecord
// @Router			/v1/runners/{runner_id}/logs [GET]
func (s *service) OtelReadLogs(ctx *gin.Context) {
	// get the otel logs in order of their timstamp (desc)
	// returns pagination details in the header
	// X-Nuon-API-Next   returns a value, a unix time in nanoseconds rn
	// X-Nuon-API-Offset accepts a value, a unix time in nanoseconds rn
	runnerID := ctx.Param("runner_id")
	jobID := ctx.DefaultQuery("job_id", "")
	jobExecutionID := ctx.DefaultQuery("job_execution_id", "")

	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get runner"))
		return
	}

	// Pagination Facts
	// parse the offset
	var before int64
	beforeStr := ctx.GetHeader("X-Nuon-API-Offset")
	if beforeStr == "" {
		// set default value
		before = time.Now().UTC().UnixNano()
	} else {
		beforeVal, err := strconv.ParseInt(beforeStr, 10, 64)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to parse pagination query params: %w", err))
			return
		}
		before = beforeVal
	}

	// read logs from chDB
	logs, headers, readerr := s.getRunnerLogs(ctx, runnerID, before, jobID, jobExecutionID)
	if readerr != nil {
		ctx.Error(fmt.Errorf("unable to read runner logs: %w", readerr))
		return
	}

	// set the header
	for key, value := range headers {
		ctx.Header(key, value)
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getRunnerLogs(ctx context.Context, runnerID string, before int64, jobID string, jobExecutionID string) ([]app.OtelLogRecord, map[string]string, error) {
	// headers
	headers := map[string]string{"Range-Units": "items"}

	// prepare a slice to hold all of the record we will be writing
	otelLogRecords := []app.OtelLogRecord{}

	// get count
	var count int64
	query := s.chDB.WithContext(ctx).Where("runner_id = ?", runnerID)
	// compose query from values from query params

	if jobID != "" {
		query = query.Where("runner_job_id = ?", jobID)
	}
	if jobExecutionID != "" {
		query = query.Where("runner_job_execution_id = ?", jobExecutionID)
	}

	// Query: get records
	res := s.chDB.WithContext(ctx).Order("timestamp desc").Limit(PageSize).Where(query).Where("toUnixTimestamp64Nano(timestamp) < ?", before).Find(&otelLogRecords)
	if res.Error != nil {
		return nil, headers, fmt.Errorf("unable to retrieve logs: %w", res.Error)
	}

	// Query: get record count
	countres := s.chDB.WithContext(ctx).Find(&app.OtelLogRecord{}).Select("id").Where(query).Count(&count)
	if countres.Error != nil {
		return nil, headers, fmt.Errorf("unable to retrieve logs: %w", countres.Error)
	}

	// determine next
	var next string
	if len(otelLogRecords) < PageSize {
		next = ""
	} else {
		last := otelLogRecords[len(otelLogRecords)-1]
		next = fmt.Sprintf("%d", last.Timestamp.UnixNano())
	}

	// add headers
	headers["X-Nuon-API-Next"] = next
	headers["count"] = strconv.FormatInt(count, 10)

	return otelLogRecords, headers, nil
}

package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const PageSize int = 100

type Page struct {
	Data  []app.OtelLogRecord `json:"data"`
	Next  string              `json:"next"`
	Count int64               `json:"count"`
}

// @ID OtelReadLogs
// @Summary	get a runner's logs
// @Description.markdown runner_otel_read_logs.md
// @Param			runner_id	path	string	true	"runner ID"
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

	// NOTE(fd): consider checking if the runner exists in the given org
	runnerID := ctx.Param("runner_id")

	// Pagination Facts
	// parse the offset
	var before int64
	beforeStr := ctx.GetHeader("X-Nuon-API-Offset")
	if beforeStr == "" {
		// set default value
		before = time.Now().UTC().UnixNano()
		fmt.Printf("before :: %+v (default)\n", before)
	} else {
		beforeVal, err := strconv.ParseInt(beforeStr, 10, 64)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to parse pagination query params: %w", err))
			return
		}
		before = beforeVal
	}

	// parse query params
	job_id := ctx.DefaultQuery("job_id", "")
	job_execution_id := ctx.DefaultQuery("job_execution_id", "")

	fmt.Printf("job_id=%s job_execution_id=%s\n", job_id, job_execution_id)

	// read logs from chDB
	logs, headers, readerr := s.getRunnerLogs(ctx, runnerID, before, job_id, job_execution_id)
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

func (s *service) getRunnerLogs(ctx context.Context, runnerID string, before int64, job_id string, job_execution_id string) ([]app.OtelLogRecord, map[string]string, error) {
	// headers
	headers := map[string]string{"Range-Units": "items"}

	// prepare a slice to hold all of the record we will be writing
	otelLogRecords := []app.OtelLogRecord{}

	// get count
	var count int64
	query := s.chDB.WithContext(ctx).Where("runner_id = ?", runnerID)
	// compose query from values from query params

	if job_id != "" {
		query = query.Where("runner_job_id = ?", job_id)
	}
	if job_execution_id != "" {
		query = query.Where("runner_job_execution_id = ?", job_execution_id)
	}

	// Query: get records
	res := s.chDB.WithContext(ctx).Order("timestamp asc").Limit(PageSize).Where(query).Where("toUnixTimestamp64Nano(timestamp) < ?", before).Find(&otelLogRecords)
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

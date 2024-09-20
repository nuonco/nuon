package service

import (
	"context"
	"fmt"
	"strconv"

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
	// get the otel logs in order of their timstamp (asc)
	// TODO(fd): add [cursor-based?] pagination
	// NOTE(fd): the pagination is mocked out but not implemented

	// NOTE(fd): consider checking if the runner exists in the given org after looking for logs
	runnerID := ctx.Param("runner_id")

	// pagination facts
	// we expect a header w/ the value for next in the header. we choose how to handle it internally
	// the current implemention will use after in a `next` header
	var after int64
	afterStr := ctx.GetHeader("X-Nuon-API-Offset")
	if afterStr == "" {
		// set default value
		after = 0
	} else {
		afterVal, err := strconv.ParseInt(afterStr, 10, 64)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to parse pagination query params: %w", err))
			return
		}
		after = afterVal
	}

	// write the logs to the db
	logs, headers, readerr := s.getRunnerLogs(ctx, runnerID, after)
	if readerr != nil {
		ctx.Error(fmt.Errorf("unable to read runner logs: %w", readerr))
		return
	}

	for key, value := range headers {
		ctx.Header(key, value)
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getRunnerLogs(ctx context.Context, runnerID string, after int64) ([]app.OtelLogRecord, map[string]string, error) {
	// headers
	headers := map[string]string{"Range-Units": "items"}
	// prepare a slice to hold all of the record we will be writing
	otelLogRecords := []app.OtelLogRecord{}

	// get count
	var count int64
	countres := s.chDB.WithContext(ctx).Find(&otelLogRecords).Where("runner_id = ?", runnerID).Count(&count)
	if countres.Error != nil {
		return nil, headers, fmt.Errorf("unable to retrieve logs: %w", countres.Error)
	}

	// get records
	res := s.chDB.WithContext(ctx).Limit(PageSize).Find(&otelLogRecords).Where("runner_id = ? AND timestamp > ?", runnerID, after).Order("timestamp asc")
	if res.Error != nil {
		return nil, headers, fmt.Errorf("unable to retrieve logs: %w", res.Error)
	}

	// determine next
	var next string
	if len(otelLogRecords) < PageSize {
		next = ""
	} else {
		last := otelLogRecords[len(otelLogRecords)-1]
		next = fmt.Sprintf("%d", last.Timestamp.UnixNano())
	}
	headers["X-Nuon-API-Next"] = next
	headers["count"] = strconv.FormatInt(count, 10)

	// parse from json

	return otelLogRecords, headers, nil
}

package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	PageSize int = 250
)

// @ID LogStreamReadLogs
// @Summary	read a log stream's logs
// @Description.markdown log_stream_read_logs.md
// @Param			log_stream_id	path	string	true	"log stream ID"
// @Param			X-Nuon-API-Offset	header	string	true	"log stream offset"
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
// @Router			/v1/log-streams/{log_stream_id}/logs [GET]
func (s *service) LogStreamReadLogs(ctx *gin.Context) {
	// get the otel logs in order of their timstamp (desc)
	// returns pagination details in the header
	// X-Nuon-API-Next   returns a value, a unix time in nanoseconds rn
	// X-Nuon-API-Offset accepts a value, a unix time in nanoseconds rn
	logStreamID := ctx.Param("log_stream_id")

	_, err := s.getLogStream(ctx, logStreamID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get runner"))
		return
	}

	// Pagination Facts
	// parse the offset
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

	// read logs from chDB
	logs, headers, readerr := s.getLogStreamLogs(ctx, logStreamID, after)
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

func (s *service) getLogStreamLogs(ctx context.Context, runnerID string, after int64) ([]app.OtelLogRecord, map[string]string, error) {
	// headers
	headers := map[string]string{"Range-Units": "items"}

	// prepare a slice to hold all of the record we will be writing
	otelLogRecords := []app.OtelLogRecord{}

	res := s.chDB.WithContext(ctx).
		Where("log_stream_id = ?", runnerID)

	if after != 0 {
		res.Where("toUnixTimestamp64Nano(timestamp) > ?", after)
	}

	res.
		Order("timestamp asc").
		Limit(PageSize).
		Find(&otelLogRecords)
	if res.Error != nil {
		return nil, headers, fmt.Errorf("unable to retrieve logs: %w", res.Error)
	}

	// Query: get total record count
	// get count
	var count int64
	countres := s.chDB.WithContext(ctx).
		Where("log_stream_id = ?", runnerID).
		Find(&[]app.OtelLogRecord{}).
		Count(&count)
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

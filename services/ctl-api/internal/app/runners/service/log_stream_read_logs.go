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
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	PageSize             int    = 100
	nestedAttributeRegex string = `^(?:[a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)?)$` // https://regex101.com/r/179bxx/1
)

// @ID						LogStreamReadLogs
// @Summary				read a log stream's logs
// @Description.markdown	log_stream_read_logs.md
// @Param					log_stream_id		path	string	true	"log stream ID"
// @Param					X-Nuon-API-Offset	header	string	false	"log stream offset"
// query param for order
// @Param					order query	string	false	"resource attribute filters"	default(asc)
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.OtelLogRecord
// @Router					/v1/log-streams/{log_stream_id}/logs [GET]
func (s *service) LogStreamReadLogs(ctx *gin.Context) {
	logStreamID := ctx.Param("log_stream_id")

	_, err := s.getLogStream(ctx, logStreamID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get runner"))
		return
	}

	// Pagination Facts
	// parse the offset
	var after int64
	order := ctx.DefaultQuery("order", "asc")
	if order != "asc" && order != "desc" {
		ctx.Error(errors.New("invalid order query parameter, must be 'asc' or 'desc'"))
		return
	}

	afterStr := ctx.GetHeader("X-Nuon-API-Offset")
	if afterStr == "" {
		// set default value based on order
		if order == "asc" {
			after = 0 // start from beginning of time
		} else {
			after = time.Now().UnixNano() // start from current time
		}
	} else {
		afterVal, err := strconv.ParseInt(afterStr, 10, 64)
		if err != nil {
			ctx.Error(errors.Wrap(err, "unable to parse pagination query params"))
			return
		}
		after = afterVal
	}

	// read logs from chDB
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to read org id from context"))
		return
	}

	logs, headers, readErr := s.getLogStreamLogs(ctx, logStreamID, orgID, after, order)
	if readErr != nil {
		ctx.Error(errors.Wrap(readErr, "unable to read runner logs"))
		return
	}

	// set the header
	for key, value := range headers {
		ctx.Header(key, value)
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getLogStreamLogs(ctx context.Context, logStreamID string, orgID string, after int64, order string) ([]app.OtelLogRecord, map[string]string, error) {
	// NOTE(jm): this is a temporary mechanism, while we test log reading. Ultimately, this should be done in a
	// middleware.
	ctx, cancelFn := context.WithTimeout(ctx, time.Second*5)
	defer cancelFn()

	// headers
	headers := map[string]string{"Range-Units": "items"}

	// prepare a slice to hold all of the record we will be writing
	otelLogRecords := []app.OtelLogRecord{}

	res := s.chDB.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where("log_stream_id = ?", logStreamID)

	// For ascending order, get records >= cursor (inclusive)
	// For descending order, get records <= cursor (inclusive)
	if order == "asc" {
		res.Where("toUnixTimestamp64Nano(timestamp) >= ?", after)
	} else {
		res.Where("toUnixTimestamp64Nano(timestamp) <= ?", after)
	}

	res.
		Order(fmt.Sprintf("timestamp %s", order)).
		Limit(PageSize).
		Find(&otelLogRecords)
	if res.Error != nil {
		return nil, headers, errors.Wrap(res.Error, "unable to retrieve logs")
	}

	// Query: get total record count
	var count int64
	countres := s.chDB.WithContext(ctx).
		Model(&app.OtelLogRecord{}).
		Where("log_stream_id = ?", logStreamID).
		Count(&count)
	if countres.Error != nil {
		return nil, headers, errors.Wrap(countres.Error, "unable to retrieve logs count")
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

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

	// Parse order parameter
	order := ctx.DefaultQuery("order", "asc")
	if order != "asc" && order != "desc" {
		ctx.Error(errors.New("invalid order query parameter, must be 'asc' or 'desc'"))
		return
	}

	// Parse cursor
	var cursor int64
	cursorStr := ctx.GetHeader("X-Nuon-API-Offset")
	if cursorStr != "" {
		cursorVal, parseErr := strconv.ParseInt(cursorStr, 10, 64)
		if parseErr != nil {
			ctx.Error(errors.Wrap(parseErr, "unable to parse pagination cursor"))
			return
		}
		cursor = cursorVal
	}

	// Read logs from chDB
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to read org id from context"))
		return
	}

	logs, headers, readErr := s.getLogStreamLogs(ctx, logStreamID, orgID, cursor, order)
	if readErr != nil {
		ctx.Error(errors.Wrap(readErr, "unable to read runner logs"))
		return
	}

	// Set headers
	for key, value := range headers {
		ctx.Header(key, value)
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getLogStreamLogs(ctx context.Context, logStreamID string, orgID string, cursor int64, order string) ([]app.OtelLogRecord, map[string]string, error) {
	ctx, cancelFn := context.WithTimeout(ctx, time.Second*5)
	defer cancelFn()

	headers := map[string]string{"Range-Units": "items"}

	// Get total count first
	var totalCount int64
	countRes := s.chDB.WithContext(ctx).
		Model(&app.OtelLogRecord{}).
		Where("org_id = ?", orgID).
		Where("log_stream_id = ?", logStreamID).
		Count(&totalCount)
	if countRes.Error != nil {
		return nil, headers, errors.Wrap(countRes.Error, "unable to retrieve logs count")
	}
	headers["count"] = strconv.FormatInt(totalCount, 10)

	// Handle empty results
	if totalCount == 0 {
		headers["X-Nuon-API-Next"] = ""
		return []app.OtelLogRecord{}, headers, nil
	}

	var otelLogRecords []app.OtelLogRecord

	if order == "asc" {
		// ASC: Standard forward pagination
		res := s.chDB.WithContext(ctx).
			Where("org_id = ?", orgID).
			Where("log_stream_id = ?", logStreamID)

		if cursor > 0 {
			res = res.Where("toUnixTimestamp64Nano(timestamp) > ?", cursor)
		}

		res.Order("timestamp ASC").
			Limit(PageSize).
			Find(&otelLogRecords)
		if res.Error != nil {
			return nil, headers, errors.Wrap(res.Error, "unable to retrieve logs")
		}

		// Determine next cursor
		if len(otelLogRecords) < PageSize {
			headers["X-Nuon-API-Next"] = ""
		} else {
			last := otelLogRecords[len(otelLogRecords)-1]
			headers["X-Nuon-API-Next"] = fmt.Sprintf("%d", last.Timestamp.UnixNano())
		}
	} else {
		// DESC: Reverse pagination using ASC query + offset calculation
		var offset int64

		if cursor == 0 {
			// First page - start from the end
			offset = totalCount - int64(PageSize)
			if offset < 0 {
				offset = 0
			}
		} else {
			// Subsequent pages - count records before cursor (timestamp < cursor)
			var countBeforeCursor int64
			countRes := s.chDB.WithContext(ctx).
				Model(&app.OtelLogRecord{}).
				Where("org_id = ?", orgID).
				Where("log_stream_id = ?", logStreamID).
				Where("toUnixTimestamp64Nano(timestamp) < ?", cursor).
				Count(&countBeforeCursor)
			if countRes.Error != nil {
				return nil, headers, errors.Wrap(countRes.Error, "unable to count remaining logs")
			}

			// Get the last PageSize records from what remains
			offset = countBeforeCursor - int64(PageSize)
			if offset < 0 {
				offset = 0
			}
		}

		// Query with ASC order and offset
		res := s.chDB.WithContext(ctx).
			Where("org_id = ?", orgID).
			Where("log_stream_id = ?", logStreamID).
			Order("timestamp ASC").
			Offset(int(offset)).
			Limit(PageSize).
			Find(&otelLogRecords)
		if res.Error != nil {
			return nil, headers, errors.Wrap(res.Error, "unable to retrieve logs")
		}

		// Reverse the results in memory
		for i, j := 0, len(otelLogRecords)-1; i < j; i, j = i+1, j-1 {
			otelLogRecords[i], otelLogRecords[j] = otelLogRecords[j], otelLogRecords[i]
		}

		// Determine next cursor (last element after reversal = oldest timestamp)
		if len(otelLogRecords) == 0 || offset == 0 {
			headers["X-Nuon-API-Next"] = ""
		} else {
			last := otelLogRecords[len(otelLogRecords)-1]
			headers["X-Nuon-API-Next"] = fmt.Sprintf("%d", last.Timestamp.UnixNano())
		}
	}

	return otelLogRecords, headers, nil
}

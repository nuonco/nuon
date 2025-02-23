package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	chhelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/ch/helpers"
)

const (
	PageSize             int    = 250
	nestedAttributeRegex string = `^(?:[a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)?)$` // https://regex101.com/r/179bxx/1
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
			ctx.Error(errors.Wrap(err, "unable to parse pagination query params"))
			return
		}
		after = afterVal
	}

	// grab any and all attribute query params
	queryParams := ctx.Request.URL.Query()
	resourceAttrQueryParams := map[string]string{}
	logAttrQueryParams := map[string]string{}
	for key, value := range queryParams {
		if strings.HasPrefix(key, "resource.") {
			newKey := strings.Replace(key, "resource.", "", 1)
			match, _ := regexp.MatchString(nestedAttributeRegex, newKey)
			if match {
				resourceAttrQueryParams[newKey] = value[0]
			}
		} else if strings.HasPrefix(key, "log.") {
			newKey := strings.Replace(key, "log.", "", 1)
			match, _ := regexp.MatchString(nestedAttributeRegex, newKey)
			if match {
				logAttrQueryParams[newKey] = value[0]
			}
		}
	}

	// read logs from chDB
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to read org id from context"))
		return
	}
	logs, headers, readErr := s.getLogStreamLogs(ctx, logStreamID, orgID, after, resourceAttrQueryParams, logAttrQueryParams)
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

func (s *service) getLogStreamLogs(ctx context.Context, logStreamID string, orgID string, after int64, resourceAttrQueryParams map[string]string, logAttrQueryParams map[string]string) ([]app.OtelLogRecord, map[string]string, error) {
	// headers
	headers := map[string]string{"Range-Units": "items"}

	// prepare a slice to hold all of the record we will be writing
	otelLogRecords := []app.OtelLogRecord{}

	res := s.chDB.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where("log_stream_id = ?", logStreamID)

	if after != 0 {
		res.Where("toUnixTimestamp64Nano(timestamp) > ?", after)
	}

	for key, value := range resourceAttrQueryParams {
		columnName := chhelpers.NestedColumnName("resource_attributes", key)
		res.Where(fmt.Sprintf("%s = ?", res.Config.NamingStrategy.ColumnName("", columnName)), value)
	}

	for key, value := range logAttrQueryParams {
		columnName := chhelpers.NestedColumnName("log_attributes", key)
		res.Where(fmt.Sprintf("%s = ?", res.Config.NamingStrategy.ColumnName("", columnName)), value)
	}

	res.
		Order("timestamp asc").
		Limit(PageSize).
		Find(&otelLogRecords)
	if res.Error != nil {
		return nil, headers, errors.Wrap(res.Error, "unable to retrieve logs")
	}

	// Query: get total record count
	// get count
	var count int64
	countres := s.chDB.WithContext(ctx).
		Where("log_stream_id = ?", logStreamID).
		Find(&[]app.OtelLogRecord{}).
		Count(&count)
	if countres.Error != nil {
		return nil, headers, errors.Wrap(countres.Error, "unable to retrieve logs: %w")
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

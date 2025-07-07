package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

type AuditLogs []app.InstallAuditLog

// @ID										GetInstallAuditLogs
// @Summary								get install audit logs
// @Description.markdown	get_install_audit_logs.md
// @Param									install_id	path	string	true	"install ID"
// @Param									start	query	string	true	"start timestamp for audit log range"
// @Param									end	query	string	true	"end timestamp for audit log range"
// @Tags									installs
// @Accept								json
// @Produce								text/csv
// @Security							APIKey
// @Security							OrgID
// @Failure								400	{object}	stderr.ErrResponse
// @Failure								401	{object}	stderr.ErrResponse
// @Failure								403	{object}	stderr.ErrResponse
// @Failure								404	{object}	stderr.ErrResponse
// @Failure								500	{object}	stderr.ErrResponse
// @Success								200	{object}	AuditLogs
// @Router								/v1/installs/{install_id}/audit_logs [get]
func (s *service) GetInstallAuditLogs(ctx *gin.Context) {
	// get install state
	installID := ctx.Param("install_id")

	startTS, err := time.Parse(time.RFC3339Nano, ctx.Query("start"))
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid start timestamp: %w", err),
			Description: "Start timestamp must be in RFC3339/Nano format (e.g., 2023-10-01T00:00:00Z).",
		})
		return
	}

	endTS, err := time.Parse(time.RFC3339Nano, ctx.Query("end"))
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid end timestamp: %w", err),
			Description: "End timestamp must be in RFC3339/Nano format (e.g., 2023-10-01T23:59:59Z).",
		})
		return
	}

	if startTS.After(endTS) {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("start timestamp cannot be after end timestamp"),
			Description: "Start timestamp must be before or equal to end timestamp.",
		})
		return
	}

	// get audit logs from the view
	auditLogs, err := s.helpers.GetInstallAuditLogs(ctx, installID, startTS, endTS)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to get install audit logs: %w", err),
			Description: "Failed to retrieve audit logs for the specified install and time range.",
		})
		return
	}

	if len(auditLogs) == 0 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("no audit logs found for install %s in the specified range", installID),
			Description: fmt.Sprintf("No audit logs found for install %s between the specified time range.", installID),
		})
		return
	}

	// convert audit logs to CSV format
	var response bytes.Buffer
	{
		var csvData [][]string
		csvData = append(csvData, []string{"Install ID", "Type", "Time Stamp", "Log Line"})
		for _, log := range auditLogs {
			csvData = append(csvData, []string{
				log.InstallID,
				log.Type,
				log.TimeStamp.Format(time.RFC3339Nano),
				log.LogLine,
			})
		}
		writer := csv.NewWriter(&response)
		err := writer.WriteAll(csvData)
		if err != nil {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("error writing CSV data: %w", err),
				Description: "Failed to write audit logs to CSV format.",
			})
			return
		}
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_audit_history.csv\"", installID))
	ctx.Data(http.StatusOK, "text/csv", response.Bytes())
}

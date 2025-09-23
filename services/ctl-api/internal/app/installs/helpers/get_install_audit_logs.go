package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

const (
	// defaultInstallAuditLogsLimit is the default limit for install audit logs.
	defaultInstallAuditLogsLimit = 10_000
)

// GetInstallAuditLogs gets the audit logs for an install from the DB.
func (h *Helpers) GetInstallAuditLogs(ctx context.Context, installID string, startTS, endTS time.Time) ([]app.InstallAuditLog, error) {
	var auditLogs []app.InstallAuditLog
	res := h.db.WithContext(ctx).
		Scopes(
			scopes.WithOverrideTable(views.DefaultViewName(h.db, &app.InstallAuditLog{}, 1)),
		).
		Order("time_stamp ASC").
		Limit(defaultInstallAuditLogsLimit).
		Find(&auditLogs, "install_id = ? AND time_stamp BETWEEN ? AND ?", installID, startTS, endTS)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			// No rows found, return an empty slice
			return auditLogs, nil
		}
		return nil, fmt.Errorf("unable to get install audit logs: %w", res.Error)
	}

	return auditLogs, nil
}

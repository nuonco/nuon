package statetelemetry

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"

var healthStates = []string{"unhealthy", "degraded", "other", "unknown", "progressing", "healthy", "not-applicable", "unset"}
var reportAgeStates = []string{"within_5m", "over_5m", "never_reported", "invalid_time"}
var sandboxStates = []string{"deleted", "delete_failed", "active", "error", "queued", "deprovisioned", "deprovisioning", "provisioning", "reprovisioning", "access_error", "unknown", "empty", "unset", "other", "missing"}

var installHealthSQL = `
SELECT i.org_id, i.app_id, i.id AS install_id, ` + viewsql.InstallHealthRollupV1("i.id") + ` AS state, i.state AS lifecycle
FROM live_installs i
`

const reportAgeSQL = `
SELECT org_id, app_id, id AS install_id,
    CASE WHEN last_health_report_at IS NULL THEN 'never_reported'
         WHEN last_health_report_at <= 'epoch'::timestamptz OR last_health_report_at > transaction_timestamp() THEN 'invalid_time'
         WHEN last_health_report_at >= transaction_timestamp() - interval '5 minutes' THEN 'within_5m'
         ELSE 'over_5m' END AS state
FROM live_installs
`

const sandboxLifecycleSQL = `
SELECT i.org_id, i.app_id, i.id AS install_id,
    CASE WHEN s.id IS NULL THEN 'missing'
         WHEN COALESCE(NULLIF(s.status_v2->>'status', ''), s.status) = 'missing' THEN 'other'
         ELSE COALESCE(NULLIF(s.status_v2->>'status', ''), s.status, '') END AS state
FROM live_installs i LEFT JOIN install_sandboxes s ON s.install_id = i.id AND s.org_id = i.org_id AND s.deleted_at = 0
`

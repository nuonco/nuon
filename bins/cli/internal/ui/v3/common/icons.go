package common

import "github.com/nuonco/nuon/sdks/nuon-go/models"

// StatusIconMap mirrors the dashboard-ui status buckets defined in
// services/dashboard-ui/client/utils/status-utils.ts so the CLI and the
// web UI render visually consistent step/workflow states.
//
// Buckets:
//   - success bucket  → "✓"
//   - error bucket    → "⊗"
//   - warn bucket     → "⚠" (cancelled is intentionally kept as "⊗" — it's
//     a terminal state and the legacy CLI glyph reads better in lists)
//   - pending/neutral → "⏲"
//   - in-progress     → "→" (replaced with a live spinner frame at render
//     time for any status in InProgressStatuses; see workflow/list_step.go)
//   - skipped         → "→"
//   - inert/brand     → "∙" (default fallback)
var StatusIconMap = map[models.AppStatus]string{
	// success bucket
	models.AppStatusSuccess:                             "✓",
	models.AppStatusApproved:                            "✓",
	models.AppStatusActive:                              "✓",
	models.AppStatusNoDashDrift:                         "✓",
	models.AppStatus(models.AppOperationStatusFinished): "✓",

	// error bucket
	models.AppStatusError: "⊗",

	// warn bucket
	models.AppStatusWarning:              "⚠",
	models.AppStatusApprovalDashAwaiting: "⚠",
	models.AppStatusApprovalDashDenied:   "⚠",
	models.AppStatusApprovalDashExpired:  "⚠",
	models.AppStatusApprovalDashRetry:    "⚠",
	models.AppStatusOutdated:             "⚠",
	models.AppStatusDrifted:              "⚠",
	models.AppStatusExpired:              "⚠",

	// terminal "stopped" — keep legacy glyph; conceptually warn but reads
	// better as a hard-stop indicator next to ✓ / ⚠ rows.
	models.AppStatusCancelled: "⊗",

	// pending / neutral bucket
	models.AppStatusPending: "⏲",
	models.AppStatusNoop:    "⏲",

	// in-progress bucket (animated → spinner frame at render time)
	models.AppStatusInDashProgress:          "→",
	models.AppStatusPlanning:                "→",
	models.AppStatusApplying:                "→",
	models.AppStatusProvisioning:            "→",
	models.AppStatusBuilding:                "→",
	models.AppStatusQueued:                  "→",
	models.AppStatusGenerating:              "→",
	models.AppStatusRetrying:                "→",
	models.AppStatusCheckingDashPlan:        "→",
	models.AppStatusAwaitingDashUserDashRun: "→",
	models.AppStatusDeleting:                "→",

	// skipped bucket
	models.AppStatusAutoDashSkipped: "→",
	models.AppStatusUserDashSkipped: "→",

	// inert / brand bucket — fall through to "∙" via GetStatusIcon:
	//   models.AppStatusNotDashAttempted, models.AppStatusDiscarded
}

// InProgressStatuses lists the AppStatus values that represent active,
// in-flight work. Renderers should swap the static glyph for an animated
// spinner frame when a step is in any of these states.
var InProgressStatuses = map[models.AppStatus]struct{}{
	models.AppStatusInDashProgress:          {},
	models.AppStatusPlanning:                {},
	models.AppStatusApplying:                {},
	models.AppStatusProvisioning:            {},
	models.AppStatusBuilding:                {},
	models.AppStatusQueued:                  {},
	models.AppStatusGenerating:              {},
	models.AppStatusRetrying:                {},
	models.AppStatusCheckingDashPlan:        {},
	models.AppStatusAwaitingDashUserDashRun: {},
	models.AppStatusDeleting:                {},
}

// IsInProgressStatus reports whether a status represents active, in-flight
// work that should be rendered with an animated spinner.
func IsInProgressStatus(status models.AppStatus) bool {
	_, ok := InProgressStatuses[status]
	return ok
}

func GetStatusIcon(status models.AppStatus) string {
	icon, ok := StatusIconMap[status]
	if !ok {
		return "∙"
	}
	return icon
}

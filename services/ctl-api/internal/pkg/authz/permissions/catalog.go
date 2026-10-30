package permissions

import (
	"fmt"
	"slices"
)

type PermissionID string

const (
	OrgRead             PermissionID = "org.read"
	OrgWrite            PermissionID = "org.write"
	AppRead             PermissionID = "app.read"
	AppWrite            PermissionID = "app.write"
	AppAccessManage     PermissionID = "app.access.manage"
	InstallRead         PermissionID = "install.read"
	InstallWrite        PermissionID = "install.write"
	InstallAccessManage PermissionID = "install.access.manage"
	RunnerRead          PermissionID = "runner.read"
	RunnerConfigRead    PermissionID = "runner.config.read"
	RunnerWrite         PermissionID = "runner.write"
	RunnerJobRead       PermissionID = "runner_job.read"
	RunnerJobConfigRead PermissionID = "runner_job.config.read"
	RunnerJobClaim      PermissionID = "runner_job.claim"
	RunnerJobWrite      PermissionID = "runner_job.write"
	TelemetryRead       PermissionID = "telemetry.read"
	TelemetryConfigRead PermissionID = "telemetry.config.read"
	TelemetryIssue      PermissionID = "telemetry.issue"
	TelemetryWrite      PermissionID = "telemetry.write"
	LogStreamRead       PermissionID = "log_stream.read"
	LogStreamWrite      PermissionID = "log_stream.write"
	StackRead           PermissionID = "stack.read"
	StackReport         PermissionID = "stack.report"
	ExecutionStateRead  PermissionID = "execution_state.read"
	ExecutionStateWrite PermissionID = "execution_state.write"
	EventSubmit         PermissionID = "event.submit"
	NotificationRead    PermissionID = "notification.read"
	NotificationWrite   PermissionID = "notification.write"
	NotificationDeliver PermissionID = "notification.deliver"
	OperationExecute    PermissionID = "operation.execute"
)

func ParsePermissionID(value string) (PermissionID, error) {
	p := PermissionID(value)
	if p.resourceKind() == "" {
		return "", fmt.Errorf("invalid permission ID %q", value)
	}
	return p, nil
}

func (p PermissionID) resourceKind() ResourceKind {
	switch p {
	case OrgRead, OrgWrite:
		return KindOrg
	case AppRead, AppWrite, AppAccessManage:
		return KindApp
	case InstallRead, InstallWrite, InstallAccessManage:
		return KindInstall
	case RunnerRead, RunnerConfigRead, RunnerWrite:
		return KindRunner
	case RunnerJobRead, RunnerJobConfigRead, RunnerJobClaim, RunnerJobWrite:
		return KindRunnerJob
	case TelemetryRead, TelemetryConfigRead, TelemetryIssue, TelemetryWrite:
		return KindTelemetry
	case LogStreamRead, LogStreamWrite:
		return KindLogStream
	case StackRead, StackReport:
		return KindStack
	case ExecutionStateRead, ExecutionStateWrite:
		return KindExecutionState
	case EventSubmit:
		return KindEvent
	case NotificationRead, NotificationWrite, NotificationDeliver:
		return KindNotification
	case OperationExecute:
		return KindOperation
	default:
		return ""
	}
}

func (p PermissionID) appliesTo(kind ResourceKind) bool {
	return p.resourceKind() != "" && (p.resourceKind() == kind ||
		(p == AppWrite && kind == KindOrg) ||
		(p == InstallWrite && kind == KindApp) ||
		(p == StackReport && kind == KindPhoneHome))
}

func (p PermissionID) customerPermission() bool {
	switch p {
	case OrgRead, OrgWrite, AppRead, AppWrite, AppAccessManage, InstallRead, InstallWrite,
		InstallAccessManage, RunnerRead, RunnerJobRead, TelemetryRead, LogStreamRead:
		return true
	default:
		return false
	}
}

type PermissionSet map[PermissionID]struct{}

func NewPermissionSet(values ...PermissionID) (PermissionSet, error) {
	set := make(PermissionSet, len(values))
	for _, value := range values {
		if _, err := ParsePermissionID(string(value)); err != nil {
			return nil, err
		}
		set[value] = struct{}{}
	}
	return set, nil
}

func (s PermissionSet) Values() []PermissionID {
	values := make([]PermissionID, 0, len(s))
	for value := range s {
		values = append(values, value)
	}
	slices.Sort(values)
	return values
}

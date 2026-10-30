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

var permissionTargets = map[PermissionID][]ResourceKind{
	OrgRead:             {KindOrg},
	OrgWrite:            {KindOrg},
	AppRead:             {KindApp},
	AppWrite:            {KindApp, KindOrg},
	AppAccessManage:     {KindApp},
	InstallRead:         {KindInstall},
	InstallWrite:        {KindInstall, KindApp},
	InstallAccessManage: {KindInstall},
	RunnerRead:          {KindRunner},
	RunnerConfigRead:    {KindRunner},
	RunnerWrite:         {KindRunner},
	RunnerJobRead:       {KindRunnerJob},
	RunnerJobConfigRead: {KindRunnerJob},
	RunnerJobClaim:      {KindRunnerJob},
	RunnerJobWrite:      {KindRunnerJob},
	TelemetryRead:       {KindTelemetry},
	TelemetryConfigRead: {KindTelemetry},
	TelemetryIssue:      {KindTelemetry},
	TelemetryWrite:      {KindTelemetry},
	LogStreamRead:       {KindLogStream},
	LogStreamWrite:      {KindLogStream},
	StackRead:           {KindStack},
	StackReport:         {KindStack, KindPhoneHome},
	ExecutionStateRead:  {KindExecutionState},
	ExecutionStateWrite: {KindExecutionState},
	EventSubmit:         {KindEvent},
	NotificationRead:    {KindNotification},
	NotificationWrite:   {KindNotification},
	NotificationDeliver: {KindNotification},
	OperationExecute:    {KindOperation},
}

func ParsePermissionID(value string) (PermissionID, error) {
	p := PermissionID(value)
	if _, ok := permissionTargets[p]; !ok {
		return "", fmt.Errorf("invalid permission ID %q", value)
	}
	return p, nil
}

func (p PermissionID) AppliesTo(kind ResourceKind) bool {
	return slices.Contains(permissionTargets[p], kind)
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

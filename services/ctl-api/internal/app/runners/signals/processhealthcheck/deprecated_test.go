package processhealthcheck

// Copy of runners/worker/activities/healthcheck_cases_test.go.

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	basemetrics "github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runneractivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

var corpusNow = time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)

// processHealthAction mirrors the decision enum in
// runners/worker/activities/healthcheck_decisions.go so the copied tables stay
// verbatim.
type processHealthAction int

const (
	processActionNoop processHealthAction = iota
	processActionShutdown
	processActionInactive
	processActionOffline
	processActionActive
)

type processHealthCase struct {
	name              string
	status            app.RunnerProcessStatus
	shutdownRequested any
	hasShutdownKey    bool
	heartbeatAge      *time.Duration
	want              processHealthAction
}

func processAge(d time.Duration) *time.Duration { return &d }

func processHealthCases() []processHealthCase {
	cases := []processHealthCase{
		{name: "shutdown requested wins over stale heartbeat", status: app.RunnerProcessStatusActive,
			shutdownRequested: true, hasShutdownKey: true, heartbeatAge: processAge(6 * time.Minute), want: processActionShutdown},
		{name: "shutdown requested wins over fresh heartbeat on offline process", status: app.RunnerProcessStatusOffline,
			shutdownRequested: true, hasShutdownKey: true, heartbeatAge: processAge(10 * time.Second), want: processActionShutdown},
		{name: "nil shutdown_requested value is ignored", status: app.RunnerProcessStatusActive,
			shutdownRequested: nil, hasShutdownKey: true, heartbeatAge: processAge(10 * time.Second), want: processActionActive},
		{name: "no heartbeat is noop", status: app.RunnerProcessStatusActive, want: processActionNoop},
		{name: "no heartbeat on offline process is noop", status: app.RunnerProcessStatusOffline, want: processActionNoop},
		{name: "10s heartbeat is active", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(10 * time.Second), want: processActionActive},
		{name: "59s heartbeat is active", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(59 * time.Second), want: processActionActive},
		{name: "60s heartbeat is offline", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(60 * time.Second), want: processActionOffline},
		{name: "61s heartbeat is offline", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(61 * time.Second), want: processActionOffline},
		{name: "299s heartbeat is offline", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(299 * time.Second), want: processActionOffline},
		{name: "300s heartbeat is inactive", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(300 * time.Second), want: processActionInactive},
		{name: "301s heartbeat is inactive", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(301 * time.Second), want: processActionInactive},
		{name: "6m heartbeat is inactive", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(6 * time.Minute), want: processActionInactive},
		{name: "offline process with fresh heartbeat flips active", status: app.RunnerProcessStatusOffline, heartbeatAge: processAge(10 * time.Second), want: processActionActive},
		{name: "offline process still silent stays offline action", status: app.RunnerProcessStatusOffline, heartbeatAge: processAge(2 * time.Minute), want: processActionOffline},
	}

	for _, status := range []app.RunnerProcessStatus{
		app.RunnerProcessStatusInactive,
		app.RunnerProcessStatusPendingShutdown,
		app.RunnerProcessStatusShuttingDown,
		app.RunnerProcessStatusShutDown,
		app.RunnerProcessStatusError,
		app.RunnerProcessStatusUnknown,
	} {
		cases = append(cases, processHealthCase{
			name:   "status " + string(status) + " is noop",
			status: status, heartbeatAge: processAge(10 * time.Second),
			want: processActionNoop,
		})
	}

	return cases
}

type versionWarningCase struct {
	name            string
	configured      string
	reported        string
	wantWarning     bool
	wantLatestEvent bool
}

func versionWarningCases() []versionWarningCase {
	return []versionWarningCase{
		{name: "versions match", configured: "1.2.3", reported: "1.2.3"},
		{name: "empty reported version", configured: "1.2.3", reported: ""},
		{name: "empty configured version", configured: "", reported: "1.2.3"},
		{name: "cloud tracks api version", configured: "cloud", reported: "1.2.3"},
		{name: "alias tag warns", configured: "stable", reported: "1.2.3", wantWarning: true},
		{name: "semver mismatch warns", configured: "1.2.4", reported: "1.2.3", wantWarning: true},
		{name: "latest configured warns and emits event", configured: "latest", reported: "1.2.3", wantWarning: true, wantLatestEvent: true},
		{name: "latest reported warns and emits event", configured: "1.2.3", reported: "latest", wantWarning: true, wantLatestEvent: true},
	}
}

type recordedProcessEffects struct {
	transitionStatus *app.RunnerProcessStatus
	transitionDesc   string
	versionWrite     *string
	greenRow         bool
	redRow           bool
	shutdownCreated  bool
	shutdownCleared  bool
	onInactive       bool
	queueStopped     bool
}

func processStatusPtr(s app.RunnerProcessStatus) *app.RunnerProcessStatus { return &s }

func stringPtr(s string) *string { return &s }

func argOf[T any](args mock.Arguments) (T, bool) {
	for i := 0; i < len(args); i++ {
		if v, ok := args.Get(i).(T); ok {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func corpusDataConverter() converter.DataConverter {
	return converter.NewCompositeDataConverter(
		signaldb.NewPayloadConverter(),
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		converter.NewJSONPayloadConverter(),
	)
}

func corpusOrgRunner(configuredVersion string) *app.Runner {
	return &app.Runner{
		ID:     "rnr_1",
		OrgID:  "org_1",
		Org:    app.Org{Name: "Example org"},
		Status: app.RunnerStatusActive,
		RunnerGroup: app.RunnerGroup{
			Type:      app.RunnerGroupTypeOrg,
			OwnerID:   "org_1",
			OwnerType: "orgs",
			Settings:  app.RunnerGroupSettings{ContainerImageTag: configuredVersion},
		},
	}
}

func corpusProcessForCase(tc processHealthCase) *app.RunnerProcess {
	metadata := map[string]any{}
	if tc.hasShutdownKey {
		metadata["shutdown_requested"] = tc.shutdownRequested
	}
	return &app.RunnerProcess{
		ID:       "rnp_1",
		RunnerID: "rnr_1",
		Type:     app.RunnerProcessTypeInstall,
		CompositeStatus: app.CompositeStatus{
			Status:   app.Status(tc.status),
			Metadata: metadata,
		},
	}
}

func runOldProcessSignal(t *testing.T, process *app.RunnerProcess, heartbeat *app.RunnerHeartBeat, configuredVersion string) recordedProcessEffects {
	t.Helper()

	v := validator.New()
	mw, err := basemetrics.New(v, basemetrics.WithDisable(true))
	require.NoError(t, err)

	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	env.SetDataConverter(corpusDataConverter())
	env.SetStartTime(corpusNow)

	sig := &Signal{RunnerID: process.RunnerID, ProcessID: process.ID}
	sig.WithParams(&signal.Params{MW: mw, V: v})

	var got recordedProcessEffects

	env.OnActivity((*runneractivities.Activities).Get, mock.Anything, mock.Anything, mock.Anything).
		Return(corpusOrgRunner(configuredVersion), nil)
	env.OnActivity((*runneractivities.Activities).GetRunnerProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(process, nil)
	env.OnActivity((*runneractivities.Activities).GetMostRecentHeartBeatByProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(heartbeat, nil)

	env.OnActivity((*runneractivities.Activities).UpdateRunnerProcessStatus, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			req, ok := argOf[runneractivities.UpdateRunnerProcessStatusRequest](args)
			require.True(t, ok)
			if warning, has := req.Metadata["version_warning"]; has {
				s, isString := warning.(string)
				require.True(t, isString)
				got.versionWrite = &s
				return
			}
			got.transitionStatus = processStatusPtr(req.Status)
			got.transitionDesc = req.StatusDescription
		}).
		Return(&app.RunnerProcess{}, nil)

	env.OnActivity((*runneractivities.Activities).CreateHealthCheck, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			req, ok := argOf[runneractivities.CreateHealthCheckRequest](args)
			require.True(t, ok)
			switch req.Status {
			case app.RunnerStatusActive:
				got.greenRow = true
			case app.RunnerStatusError:
				got.redRow = true
			}
		}).
		Return(&app.RunnerHealthCheck{}, nil)

	env.OnActivity((*runneractivities.Activities).CreateRunnerProcessShutdown, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { got.shutdownCreated = true }).
		Return(&app.RunnerProcessShutdown{}, nil)
	env.OnActivity((*runneractivities.Activities).ClearProcessShutdownRequested, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { got.shutdownCleared = true }).
		Return(nil)
	env.OnActivity(new(sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { got.onInactive = true }).
		Return(&sharedactivities.EnqueueSignalToOwnerResponse{}, nil)
	env.OnActivity((*runneractivities.Activities).StopProcessQueue, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { got.queueStopped = true }).
		Return(nil)

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	return got
}

func TestOldProcessHealthcheckCorpus(t *testing.T) {
	for _, tc := range processHealthCases() {
		t.Run(tc.name, func(t *testing.T) {
			process := corpusProcessForCase(tc)
			var heartbeat *app.RunnerHeartBeat
			if tc.heartbeatAge != nil {
				heartbeat = &app.RunnerHeartBeat{
					RunnerID:  process.RunnerID,
					ProcessID: process.ID,
					Version:   "1.2.3",
					CreatedAt: corpusNow.Add(-*tc.heartbeatAge),
				}
			}

			got := runOldProcessSignal(t, process, heartbeat, "1.2.3")

			var want recordedProcessEffects
			switch tc.want {
			case processActionNoop:
			case processActionShutdown:
				want.shutdownCreated = true
				want.shutdownCleared = true
			case processActionInactive:
				want.transitionStatus = processStatusPtr(app.RunnerProcessStatusInactive)
				want.transitionDesc = "no heartbeat received for 5 minutes"
				want.onInactive = true
				want.queueStopped = true
			case processActionOffline:
				if tc.status != app.RunnerProcessStatusOffline {
					want.transitionStatus = processStatusPtr(app.RunnerProcessStatusOffline)
					want.transitionDesc = "Runner is offline and will be marked inactive in 5 minutes"
				}
				want.redRow = true
			case processActionActive:
				if tc.status == app.RunnerProcessStatusOffline {
					want.transitionStatus = processStatusPtr(app.RunnerProcessStatusActive)
					want.transitionDesc = "heartbeat received"
				}
				want.greenRow = true
				// Old-only divergence: legacy path writes version_warning
				// unconditionally every tick; the batch port writes it only on
				// change (matching versions here → empty warning value).
				want.versionWrite = stringPtr("")
			}
			require.Equal(t, want, got)
		})
	}
}

func TestOldProcessVersionWarningCorpus(t *testing.T) {
	fresh := 10 * time.Second

	for _, tc := range versionWarningCases() {
		t.Run(tc.name, func(t *testing.T) {
			process := corpusProcessForCase(processHealthCase{status: app.RunnerProcessStatusActive})
			heartbeat := &app.RunnerHeartBeat{
				RunnerID:  process.RunnerID,
				ProcessID: process.ID,
				Version:   tc.reported,
				CreatedAt: corpusNow.Add(-fresh),
			}

			got := runOldProcessSignal(t, process, heartbeat, tc.configured)

			require.True(t, got.greenRow)
			if tc.reported == "" {
				// Old code skips checkVersionMismatch entirely on empty version.
				require.Nil(t, got.versionWrite)
				return
			}
			require.NotNil(t, got.versionWrite)
			require.Equal(t, tc.wantWarning, *got.versionWrite != "", *got.versionWrite)
		})
	}
}

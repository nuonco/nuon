package runnerhealthcheck

// Copy of runners/worker/activities/healthcheck_cases_test.go.

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	basemetrics "github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runneractivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

var corpusNow = time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)

type runnerHealthWant struct {
	result         string
	reason         string
	setMissingMng  *bool
	armOfflineTS   bool
	clearOfflineTS bool
	legacyStatus   *app.RunnerStatus
	v2Status       *app.RunnerStatus
	alert          bool
	alertOfflineAt int64
}

type runnerHealthCase struct {
	name          string
	groupType     app.RunnerGroupType
	status        app.RunnerStatus
	v2Status      app.RunnerStatus
	metadata      map[string]any
	activeBuild   bool
	activeInstall bool
	activeMng     bool
	mngChecked    bool
	want          runnerHealthWant
}

func runnerStatusPtr(s app.RunnerStatus) *app.RunnerStatus { return &s }

func runnerHealthCases() []runnerHealthCase {
	offlineFresh := corpusNow.Add(-15*time.Minute + time.Second).Unix()
	offlineExact := corpusNow.Add(-15 * time.Minute).Unix()
	offlineStale := corpusNow.Add(-time.Hour).Unix()

	healthyOrg := runnerHealthWant{result: "healthy", reason: "runner healthy"}
	unhealthyOrgReason := "no active build process"
	unhealthyInstallReason := "no active install process"

	cases := []runnerHealthCase{
		{
			name: "unknown group type does nothing", groupType: "", status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			want: runnerHealthWant{},
		},
		{
			name: "org healthy already active writes nothing", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive, activeBuild: true,
			want: healthyOrg,
		},
		{
			name: "org healthy clears stale offline_ts without resetting", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive, activeBuild: true,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want:     runnerHealthWant{result: "healthy", reason: "runner healthy", clearOfflineTS: true},
		},
		{
			name: "org recovery flips both statuses and clears offline_ts", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline, activeBuild: true,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want: runnerHealthWant{
				result: "healthy", reason: "runner healthy", clearOfflineTS: true,
				legacyStatus: runnerStatusPtr(app.RunnerStatusActive), v2Status: runnerStatusPtr(app.RunnerStatusActive),
			},
		},
		{
			name: "org first failed check arms and transitions without alert", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyOrgReason, armOfflineTS: true,
				legacyStatus: runnerStatusPtr(app.RunnerStatusOffline), v2Status: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "org unhealthy with existing offline_ts still transitioning does not re-arm", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyOrgReason,
				legacyStatus: runnerStatusPtr(app.RunnerStatusOffline), v2Status: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "org offline under alert delay does nothing", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineFresh},
			want:     runnerHealthWant{result: "unhealthy", reason: unhealthyOrgReason},
		},
		{
			name: "org offline exactly at alert delay alerts", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineExact},
			want:     runnerHealthWant{result: "unhealthy", reason: unhealthyOrgReason, alert: true, alertOfflineAt: offlineExact},
		},
		{
			name: "org offline past alert delay alerts with persisted offlineAt", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want:     runnerHealthWant{result: "unhealthy", reason: unhealthyOrgReason, alert: true, alertOfflineAt: offlineStale},
		},
		{
			name: "org offline without timestamp arms only", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			want: runnerHealthWant{result: "unhealthy", reason: unhealthyOrgReason, armOfflineTS: true},
		},
		{
			name: "legacy offline but v2 stale repairs v2 only", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusActive,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyOrgReason,
				v2Status: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "v2 offline but legacy stale repairs legacy only", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusOffline,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyOrgReason,
				legacyStatus: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "install healthy with mng writes missing_mng=false when metadata absent", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			activeInstall: true, activeMng: true, mngChecked: true,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy", setMissingMng: generics.ToPtr(false)},
		},
		{
			name: "install healthy mng metadata unchanged writes nothing", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			metadata:      map[string]any{"missing_mng_process": false},
			activeInstall: true, activeMng: true, mngChecked: true,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy"},
		},
		{
			name: "install healthy mng went missing writes flip", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			metadata:      map[string]any{"missing_mng_process": false},
			activeInstall: true, activeMng: false, mngChecked: true,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy", setMissingMng: generics.ToPtr(true)},
		},
		{
			name: "install healthy non-bool mng metadata is rewritten", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			metadata:      map[string]any{"missing_mng_process": "yes"},
			activeInstall: true, activeMng: true, mngChecked: true,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy", setMissingMng: generics.ToPtr(false)},
		},
		{
			name: "install mng unchecked never writes mng metadata", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			activeInstall: true, activeMng: false, mngChecked: false,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy"},
		},
		{
			name: "install missing install process is unhealthy", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			activeMng: true, mngChecked: true,
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyInstallReason, armOfflineTS: true,
				setMissingMng: generics.ToPtr(false),
				legacyStatus:  runnerStatusPtr(app.RunnerStatusOffline), v2Status: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "install offline past delay alerts with install reason", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			metadata:  map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale, "missing_mng_process": true},
			activeMng: false, mngChecked: true,
			want: runnerHealthWant{result: "unhealthy", reason: unhealthyInstallReason, alert: true, alertOfflineAt: offlineStale},
		},
	}

	for _, status := range []app.RunnerStatus{
		app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusPending,
	} {
		cases = append(cases, runnerHealthCase{
			name:      "skippable status " + string(status),
			groupType: app.RunnerGroupTypeOrg,
			status:    status, v2Status: status,
			activeBuild: true,
			want:        runnerHealthWant{result: "skipped"},
		})
	}

	return cases
}

// recordedRunnerEffects is what the old signal's activity calls are translated
// into; comparable to runnerHealthWant minus the metric-only fields.
type recordedRunnerEffects struct {
	setMissingMng  *bool
	armOfflineTS   bool
	clearOfflineTS bool
	legacyStatus   *app.RunnerStatus
	legacyReason   string
	v2Status       *app.RunnerStatus
	v2Reason       string
	alert          bool
	alertKey       string
}

func argOf[T any](args mock.Arguments) (T, bool) {
	for i := 0; i < len(args); i++ {
		if v, ok := args.Get(i).(T); ok {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func notFoundErr() error {
	return dbgenerics.TemporalGormError(gorm.ErrRecordNotFound, "no rows")
}

func activeProcess(processType app.RunnerProcessType) *app.RunnerProcess {
	return &app.RunnerProcess{
		ID:   "rnp_" + string(processType),
		Type: processType,
		CompositeStatus: app.CompositeStatus{
			Status: app.Status(app.RunnerProcessStatusActive),
		},
	}
}

func corpusRunnerForCase(tc runnerHealthCase) *app.Runner {
	ownerType := "orgs"
	ownerID := "org_1"
	if tc.groupType == app.RunnerGroupTypeInstall {
		ownerType = "installs"
		ownerID = "ins_1"
	}
	return &app.Runner{
		ID:          "rnr_1",
		DisplayName: "Corpus runner",
		OrgID:       "org_1",
		Org:         app.Org{Name: "Example org"},
		Status:      tc.status,
		StatusV2: app.CompositeStatus{
			Status:   app.Status(tc.v2Status),
			Metadata: tc.metadata,
		},
		RunnerGroupID: "rng_1",
		RunnerGroup: app.RunnerGroup{
			Type:      tc.groupType,
			OwnerID:   ownerID,
			OwnerType: ownerType,
		},
	}
}

func mockProcessPresence(env *testsuite.TestWorkflowEnvironment, processType app.RunnerProcessType, active bool, checked bool) {
	call := env.OnActivity((*runneractivities.Activities).GetCurrentRunnerProcess,
		mock.MatchedBy(func(req runneractivities.GetCurrentRunnerProcessRequest) bool {
			return req.ProcessType == string(processType)
		}))
	switch {
	case !checked:
		call.Return(nil, fmt.Errorf("transient lookup failure"))
	case active:
		call.Return(activeProcess(processType), nil)
	default:
		call.Return(nil, notFoundErr())
	}
}

func TestOldRunnerHealthcheckCorpus(t *testing.T) {
	v := validator.New()
	mw, err := basemetrics.New(v, basemetrics.WithDisable(true))
	require.NoError(t, err)

	for _, tc := range runnerHealthCases() {
		t.Run(tc.name, func(t *testing.T) {
			var workflowSuite testsuite.WorkflowTestSuite
			env := workflowSuite.NewTestWorkflowEnvironment()
			env.SetDataConverter(signalDataConverter())
			env.SetStartTime(corpusNow)

			runner := corpusRunnerForCase(tc)
			sig := &Signal{RunnerID: runner.ID}
			sig.WithParams(&signal.Params{MW: mw, V: v})

			var got recordedRunnerEffects

			env.OnActivity((*runneractivities.Activities).Get, mock.Anything, mock.Anything, mock.Anything).
				Return(runner, nil)

			switch tc.groupType {
			case app.RunnerGroupTypeOrg:
				mockProcessPresence(env, app.RunnerProcessTypeBuild, tc.activeBuild, true)
			case app.RunnerGroupTypeInstall:
				mockProcessPresence(env, app.RunnerProcessTypeInstall, tc.activeInstall, true)
				mockProcessPresence(env, app.RunnerProcessTypeMng, tc.activeMng, tc.mngChecked)
			}

			env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2Metadata,
				mock.MatchedBy(func(req statusactivities.UpdateRunnerStatusV2MetadataRequest) bool { return true })).
				Run(func(args mock.Arguments) {
					req, ok := argOf[statusactivities.UpdateRunnerStatusV2MetadataRequest](args)
					require.True(t, ok)
					if val, has := req.Metadata["missing_mng_process"]; has {
						b, isBool := val.(bool)
						require.True(t, isBool)
						got.setMissingMng = &b
					}
					if val, has := req.Metadata[app.RunnerOfflineTSMetadataKey]; has {
						if val == nil {
							got.clearOfflineTS = true
						} else {
							got.armOfflineTS = true
						}
					}
				}).
				Return(nil)

			env.OnActivity((*runneractivities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					req, ok := argOf[runneractivities.UpdateStatusRequest](args)
					require.True(t, ok)
					got.legacyStatus = runnerStatusPtr(req.Status)
					got.legacyReason = req.StatusDescription
				}).
				Return(nil)

			env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					req, ok := argOf[statusactivities.UpdateRunnerStatusV2Request](args)
					require.True(t, ok)
					got.v2Status = runnerStatusPtr(req.Status)
					got.v2Reason = req.StatusDescription
				}).
				Return(nil)

			env.OnActivity(new(sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					req, ok := argOf[*sharedactivities.EnqueueSignalToOwnerRequest](args)
					require.True(t, ok)
					got.alert = true
					got.alertKey = req.IdempotencyKey
				}).
				Return(&sharedactivities.EnqueueSignalToOwnerResponse{Deduplicated: true}, nil)

			if tc.groupType == app.RunnerGroupTypeInstall {
				env.OnActivity((*runneractivities.Activities).GetInstall, mock.Anything, mock.Anything, mock.Anything).
					Return(&app.Install{ID: "ins_1", Name: "install one"}, nil)
			}

			env.ExecuteWorkflow(func(ctx workflow.Context) error {
				return sig.Execute(ctx)
			})

			require.True(t, env.IsWorkflowCompleted())
			require.NoError(t, env.GetWorkflowError())

			want := recordedRunnerEffects{
				setMissingMng:  tc.want.setMissingMng,
				armOfflineTS:   tc.want.armOfflineTS,
				clearOfflineTS: tc.want.clearOfflineTS,
				legacyStatus:   tc.want.legacyStatus,
				v2Status:       tc.want.v2Status,
				alert:          tc.want.alert,
			}
			if tc.want.legacyStatus != nil {
				want.legacyReason = tc.want.reason
			}
			if tc.want.v2Status != nil {
				want.v2Reason = tc.want.reason
			}
			if tc.want.alert {
				want.alertKey = fmt.Sprintf("runner-unhealthy:%s:%d", runner.ID, tc.want.alertOfflineAt)
			}
			require.Equal(t, want, got)
		})
	}
}

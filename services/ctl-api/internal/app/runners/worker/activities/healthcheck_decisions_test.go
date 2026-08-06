package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func corpusRunner(tc runnerHealthCase) *app.Runner {
	return &app.Runner{
		ID:     "rnrtest",
		Status: tc.status,
		StatusV2: app.CompositeStatus{
			Status:   app.Status(tc.v2Status),
			Metadata: tc.metadata,
		},
		RunnerGroup: app.RunnerGroup{Type: tc.groupType},
	}
}

func corpusProcess(tc processHealthCase) (*app.RunnerProcess, *time.Time) {
	metadata := map[string]any{}
	if tc.hasShutdownKey {
		metadata["shutdown_requested"] = tc.shutdownRequested
	}
	process := &app.RunnerProcess{
		ID: "rnptest",
		CompositeStatus: app.CompositeStatus{
			Status:   app.Status(tc.status),
			Metadata: metadata,
		},
	}
	var hbAt *time.Time
	if tc.heartbeatAge != nil {
		at := corpusNow.Add(-*tc.heartbeatAge)
		hbAt = &at
	}
	return process, hbAt
}

func decisionToWant(d runnerHealthDecision) runnerHealthWant {
	want := runnerHealthWant{
		result:         d.Result,
		reason:         d.Reason,
		setMissingMng:  d.SetMissingMng,
		armOfflineTS:   d.SetOfflineTS,
		clearOfflineTS: d.ClearOfflineTS,
		alert:          d.Alert,
	}
	if d.UpdateLegacy {
		want.legacyStatus = runnerStatusPtr(d.TargetStatus)
	}
	if d.UpdateV2 {
		want.v2Status = runnerStatusPtr(d.TargetStatus)
	}
	if d.Alert {
		want.alertOfflineAt = d.AlertOfflineAt.Unix()
	}
	return want
}

func TestDecideRunnerHealthCorpus(t *testing.T) {
	for _, tc := range runnerHealthCases() {
		t.Run(tc.name, func(t *testing.T) {
			d := decideRunnerHealth(corpusNow, corpusRunner(tc), runnerProcessPresence{
				HasActiveBuild:   tc.activeBuild,
				HasActiveInstall: tc.activeInstall,
				HasActiveMng:     tc.activeMng,
				MngChecked:       tc.mngChecked,
			})
			got := decisionToWant(d)
			if got.result == "unhealthy" || got.result == "healthy" {
				require.NotZero(t, d.TargetStatus)
			}
			want := tc.want
			if want.result == "unhealthy" || want.result == "healthy" {
				// reason is always populated on evaluated runners
				require.NotEmpty(t, want.reason)
			}
			require.Equal(t, want, got)
		})
	}
}

func TestDecideProcessHealthCorpus(t *testing.T) {
	for _, tc := range processHealthCases() {
		t.Run(tc.name, func(t *testing.T) {
			process, hbAt := corpusProcess(tc)
			require.Equal(t, tc.want, decideProcessHealth(corpusNow, process, hbAt))
		})
	}
}

func TestDecideVersionWarningCorpus(t *testing.T) {
	for _, tc := range versionWarningCases() {
		t.Run(tc.name, func(t *testing.T) {
			warning, event := decideVersionWarning(tc.configured, tc.reported)
			require.Equal(t, tc.wantWarning, warning != "", warning)
			require.Equal(t, tc.wantLatestEvent, event)
		})
	}
}

package activities_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func TestInstallStateTimezones(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	if os.Getenv("AWS_ENDPOINT_URL_S3") == "" {
		t.Skip("set AWS_ENDPOINT_URL_S3 to the local nctl S3 service")
	}
	executable, err := os.Executable()
	require.NoError(t, err)
	for _, zone := range []string{"UTC", "Asia/Kolkata", "America/New_York"} {
		t.Run(zone, func(t *testing.T) {
			cmd := exec.Command(executable, "-test.run=^TestInstallStateIntegrationSuite$/^TestMarkerTimestampsPreserveInstants$", "-test.v", "-test.timeout=2m")
			for _, env := range os.Environ() {
				if !strings.HasPrefix(env, "TZ=") {
					cmd.Env = append(cmd.Env, env)
				}
			}
			cmd.Env = append(cmd.Env, "TZ="+zone)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "TZ=%s\n%s", zone, output)
		})
	}
}

func (s *InstallStateIntegrationSuite) TestMarkerTimestampsPreserveInstants() {
	for _, mark := range []struct {
		name string
		run  func() error
	}{
		{name: "whole_state", run: func() error { return s.deps.Helpers.MarkInstallStateStale(s.ctx, s.install.ID) }},
		{name: "named_partial", run: func() error {
			return s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialInputs)
		}},
		{name: "carried_partial", run: func() error {
			if err := s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialInputs); err != nil {
				return err
			}
			baseline, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
			if err != nil {
				return err
			}
			_, err = s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
				State: baseline, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
				GeneratedBy:       app.InstallStateGenerateSourceStateManager,
				RefreshedPartials: []pkgstate.PartialName{pkgstate.PartialComponents},
				RefreshStartedAt:  time.Now().UTC().Add(-24 * time.Hour),
			})
			return err
		}},
	} {
		_, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
		require.NoError(s.T(), err)
		var before, after time.Time
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Raw("SELECT clock_timestamp()").Scan(&before).Error)
		require.NoError(s.T(), mark.run())
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Raw("SELECT clock_timestamp()").Scan(&after).Error)
		var latest app.InstallStateHead
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
			Where(app.InstallStateHead{InstallID: s.install.ID}).First(&latest).Error)
		require.False(s.T(), latest.StaleAt.Empty())
		assert.True(s.T(), !latest.StaleAt.Time.Before(before) && !latest.StaleAt.Time.After(after),
			"%s marker %s must represent the actual invalidation instant between %s and %s in TZ=%s",
			mark.name, latest.StaleAt.Time, before, after, os.Getenv("TZ"))
	}
}

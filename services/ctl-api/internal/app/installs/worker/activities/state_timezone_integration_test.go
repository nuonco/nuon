package activities_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
			cmd := exec.Command(executable, "-test.run=^TestInstallStateIntegrationSuite$/^TestHeadStaleAtPreservesInstant$", "-test.v", "-test.timeout=2m")
			for _, env := range os.Environ() {
				if !strings.HasPrefix(env, "TZ=") {
					cmd.Env = append(cmd.Env, env)
				}
			}
			cmd.Env = append(cmd.Env, "TZ="+zone)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "TZ=%s\n%s", zone, output)
			require.Contains(t, string(output), "--- PASS: TestInstallStateIntegrationSuite/TestHeadStaleAtPreservesInstant")
		})
	}
}

func (s *InstallStateIntegrationSuite) TestHeadStaleAtPreservesInstant() {
	var before, after time.Time
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Raw("SELECT clock_timestamp()").Scan(&before).Error)
	s.updateInput()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Raw("SELECT clock_timestamp()").Scan(&after).Error)
	head := s.head()
	require.False(s.T(), head.StaleAt.Empty())
	require.True(s.T(), !head.StaleAt.Time.Before(before) && !head.StaleAt.Time.After(after),
		"marker %s must represent the invalidation instant between %s and %s in TZ=%s", head.StaleAt.Time, before, after, os.Getenv("TZ"))
	stale := s.baseline().State.StaleAt
	require.NotNil(s.T(), stale)
	require.True(s.T(), stale.Equal(head.StaleAt.Time))
	require.Equal(s.T(), "eu-west-1", s.fresh().Inputs.Inputs["region"])
	require.True(s.T(), s.head().StaleAt.Empty())
}

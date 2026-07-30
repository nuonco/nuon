package installs

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

// Health prints the fleet-wide component health rollup. With --output agent
// (or json), a canary or bake-period script can poll it and gate on all_healthy.
func (s *Service) Health(ctx context.Context, appID, labels string, asJSON bool) error {
	view := ui.NewListView()

	resp, err := s.api.GetInstallsHealth(ctx, appID, labels)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(resp)
		return nil
	}

	data := [][]string{
		{
			"INSTALL ID",
			"NAME",
			"HEALTH",
			"UNHEALTHY",
			"DEGRADED",
			"DETAIL",
		},
	}
	for _, inst := range resp.Installs {
		if inst == nil {
			continue
		}
		health := inst.Health
		if health == "" {
			health = "-"
		}
		data = append(data, []string{
			inst.InstallID,
			inst.InstallName,
			health,
			strconv.FormatInt(inst.UnhealthyComponents, 10),
			strconv.FormatInt(inst.DegradedComponents, 10),
			inst.HealthDescription,
		})
	}

	view.RenderTotal(data, len(resp.Installs))

	summary := "all healthy: " +
		strconv.FormatInt(resp.Healthy, 10) + " healthy, " +
		strconv.FormatInt(resp.Degraded, 10) + " degraded, " +
		strconv.FormatInt(resp.Unhealthy, 10) + " unhealthy, " +
		strconv.FormatInt(resp.Unknown, 10) + " unknown, " +
		strconv.FormatInt(resp.Unset, 10) + " not evaluated"
	if resp.AllHealthy {
		ui.PrintSuccess(summary)
		return nil
	}
	ui.PrintWarning("not " + summary)
	return nil
}

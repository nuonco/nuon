package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) RunnersList(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	rg, err := s.api.GetInstallRunnerGroup(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(rg.Runners)
		return nil
	}

	data := [][]string{
		{"ID", "NAME", "DISPLAY NAME", "PLATFORM", "STATUS", "TAINTED", "LEADER", "CREATED AT"},
	}
	for _, runner := range rg.Runners {
		leader := ""
		if runner.Leader {
			leader = "✓"
		}
		tainted := ""
		if runner.Tainted {
			tainted = "✓"
		}
		data = append(data, []string{
			runner.ID,
			runner.Name,
			runner.DisplayName,
			string(runner.Platform),
			runner.Status,
			tainted,
			leader,
			runner.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) RunnersElectLeader(ctx context.Context, installID, runnerID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	rg, err := s.api.GetInstallRunnerGroup(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	runner, err := s.api.UpdateRunnerGroupLeader(ctx, rg.ID, runnerID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(runner)
		return nil
	}

	fmt.Printf("Runner %s (%s) elected as leader for runner group %s\n", runner.ID, runner.Name, rg.ID)
	return nil
}

func (s *Service) RunnersTaint(ctx context.Context, installID, runnerID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	runner, err := s.api.TaintRunner(ctx, runnerID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(runner)
		return nil
	}

	fmt.Printf("Runner %s tainted successfully\n", runner.ID)
	return nil
}

func (s *Service) RunnersUntaint(ctx context.Context, installID, runnerID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	runner, err := s.api.UntaintRunner(ctx, runnerID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(runner)
		return nil
	}

	fmt.Printf("Runner %s untainted successfully\n", runner.ID)
	return nil
}

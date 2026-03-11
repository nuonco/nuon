package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	nuon "github.com/nuonco/nuon/sdks/nuon-go"
)

func (s *Service) getRunnerID(ctx context.Context, installID string) (string, error) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return "", err
	}

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return "", err
	}

	if install.RunnerID == "" {
		return "", fmt.Errorf("install %s does not have a runner", installID)
	}

	return install.RunnerID, nil
}

func (s *Service) RunnerGet(ctx context.Context, installID string, asJSON bool) error {
	installID, err := s.selectInstallID(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	runnerGroup, err := s.api.GetInstallRunnerGroup(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if len(runnerGroup.Runners) == 0 {
		return ui.PrintError(fmt.Errorf("install %s does not have a runner", installID))
	}

	runner := runnerGroup.Runners[0]

	if asJSON {
		ui.PrintJSON(runner)
		return nil
	}

	view := ui.NewGetView()
	fields := [][]string{
		{"id", runner.ID},
		{"name", runner.Name},
		{"status", runner.Status},
		{"status description", runner.StatusDescription},
		{"created at", runner.CreatedAt},
		{"updated at", runner.UpdatedAt},
	}
	view.Render(fields)
	return nil
}

func (s *Service) wrapRunnerMngErr(err error, action string) error {
	if nuon.IsNotFound(err) {
		return fmt.Errorf("runner not found: unable to %s", action)
	}
	if nuon.IsBadRequest(err) {
		return fmt.Errorf("unable to %s runner: bad request", action)
	}
	if nuon.IsForbidden(err) {
		return fmt.Errorf("unable to %s runner: forbidden", action)
	}
	return fmt.Errorf("unable to %s runner: %w", action, err)
}

func (s *Service) debugRunnerMng(installID, runnerID, endpoint string) {
	ui.PrintDebug(fmt.Sprintf("install_id=%s runner_id=%s endpoint=POST /v1/runners/%s/mng/%s", installID, runnerID, runnerID, endpoint))
}

func (s *Service) RunnerRestart(ctx context.Context, installID string) error {
	runnerID, err := s.getRunnerID(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	s.debugRunnerMng(installID, runnerID, "restart")
	if err := s.api.RunnerMngRestart(ctx, runnerID); err != nil {
		return ui.PrintError(s.wrapRunnerMngErr(err, "restart"))
	}

	ui.PrintLn("successfully triggered runner restart")
	return nil
}

func (s *Service) RunnerShutDown(ctx context.Context, installID string) error {
	runnerID, err := s.getRunnerID(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	s.debugRunnerMng(installID, runnerID, "shutdown")
	if err := s.api.RunnerMngShutDown(ctx, runnerID); err != nil {
		return ui.PrintError(s.wrapRunnerMngErr(err, "shut down"))
	}

	ui.PrintLn("successfully triggered runner shutdown")
	return nil
}

func (s *Service) RunnerVMShutDown(ctx context.Context, installID string) error {
	runnerID, err := s.getRunnerID(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	s.debugRunnerMng(installID, runnerID, "shutdown-vm")
	if err := s.api.RunnerMngVMShutDown(ctx, runnerID); err != nil {
		return ui.PrintError(s.wrapRunnerMngErr(err, "shut down VM for"))
	}

	ui.PrintLn("successfully triggered runner VM shutdown")
	return nil
}

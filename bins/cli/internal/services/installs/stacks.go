package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) StacksList(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	stack, err := s.api.GetInstallStack(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(stack.Versions)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"STATUS",
			"APP CONFIG ID",
			"CREATED AT",
			"UPDATED AT",
		},
	}
	for _, v := range stack.Versions {
		status := ""
		if v.CompositeStatus != nil {
			status = string(v.CompositeStatus.Status)
		}
		data = append(data, []string{
			v.ID,
			status,
			v.AppConfigID,
			v.CreatedAt,
			v.UpdatedAt,
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) StacksGet(ctx context.Context, stackID string, asJSON bool) error {
	view := ui.NewGetView()

	stack, err := s.api.GetInstallStackByID(ctx, stackID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(stack)
		return nil
	}

	fields := [][]string{
		{"id", stack.ID},
		{"install id", stack.InstallID},
		{"org id", stack.OrgID},
		{"created at", stack.CreatedAt},
		{"updated at", stack.UpdatedAt},
	}
	view.Render(fields)
	return nil
}

func (s *Service) StacksLatest(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	stack, err := s.api.GetInstallStack(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if len(stack.Versions) == 0 {
		view.Print("no stack versions found")
		return nil
	}

	latest := stack.Versions[0]

	if asJSON {
		ui.PrintJSON(latest)
		return nil
	}

	status := ""
	if latest.CompositeStatus != nil {
		status = string(latest.CompositeStatus.Status)
	}

	fields := [][]string{
		{"id", latest.ID},
		{"install stack id", latest.InstallStackID},
		{"status", status},
		{"app config id", latest.AppConfigID},
		{"created at", latest.CreatedAt},
		{"updated at", latest.UpdatedAt},
	}
	view.Render(fields)
	return nil
}

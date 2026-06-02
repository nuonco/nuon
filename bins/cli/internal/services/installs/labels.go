package installs

import (
	"context"
	"fmt"
	"sort"

	"github.com/nuonco/nuon/pkg/cli/styles"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/services/labels"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

// LabelsList prints the labels on an install.
func (s *Service) LabelsList(ctx context.Context, installIDOrName string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installIDOrName)
	if err != nil {
		return ui.PrintError(err)
	}

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		ui.PrintJSON(map[string]any{"id": install.ID, "labels": install.Labels})
		return nil
	}

	renderLabels(install.ID, install.Labels)
	return nil
}

// LabelsSet adds or overwrites labels on an install. Args are kubectl-style
// "key=value" pairs.
func (s *Service) LabelsSet(ctx context.Context, installIDOrName string, args []string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installIDOrName)
	if err != nil {
		return ui.PrintError(err)
	}

	set, remove, err := labels.ParseArgs(args)
	if err != nil {
		return ui.PrintError(err)
	}
	if len(remove) > 0 {
		return ui.PrintError(fmt.Errorf("use 'nuon installs labels unset' to remove labels"))
	}
	if len(set) == 0 {
		return ui.PrintError(fmt.Errorf("provide at least one label as key=value"))
	}

	if _, err := s.api.AddInstallLabels(ctx, installID, set); err != nil {
		return ui.PrintError(fmt.Errorf("unable to set labels: %w", err))
	}

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		ui.PrintJSON(map[string]any{"id": install.ID, "labels": install.Labels})
		return nil
	}

	ui.PrintSuccess(fmt.Sprintf("install/%s labeled", install.ID))
	renderLabels(install.ID, install.Labels)
	return nil
}

// LabelsUnset removes labels from an install by key.
func (s *Service) LabelsUnset(ctx context.Context, installIDOrName string, args []string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installIDOrName)
	if err != nil {
		return ui.PrintError(err)
	}

	keys, err := labels.ParseKeys(args)
	if err != nil {
		return ui.PrintError(err)
	}
	if len(keys) == 0 {
		return ui.PrintError(fmt.Errorf("provide at least one label key to unset"))
	}

	if _, err := s.api.RemoveInstallLabels(ctx, installID, keys); err != nil {
		return ui.PrintError(fmt.Errorf("unable to unset labels: %w", err))
	}

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		ui.PrintJSON(map[string]any{"id": install.ID, "labels": install.Labels})
		return nil
	}

	ui.PrintSuccess(fmt.Sprintf("install/%s labels removed", install.ID))
	renderLabels(install.ID, install.Labels)
	return nil
}

// renderLabels prints an install's labels as a styled KEY/VALUE table.
func renderLabels(id string, lbls map[string]string) {
	if len(lbls) == 0 {
		fmt.Printf("%s %s\n", styles.TextBold.Render(id), styles.TextSubtle.Render("(no labels)"))
		return
	}

	keys := make([]string, 0, len(lbls))
	for k := range lbls {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	view := ui.NewListView()
	data := [][]string{{"KEY", "VALUE"}}
	for _, k := range keys {
		data = append(data, []string{k, lbls[k]})
	}
	view.Render(data)
}

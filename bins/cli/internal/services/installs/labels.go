package installs

import (
	"context"
	"fmt"
	"sort"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/services/labels"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

// Label adds or removes labels on an install. Pass kubectl-style args:
// "key=value" to set, "key-" to remove. With no args, prints current labels.
func (s *Service) Label(ctx context.Context, installIDOrName string, args []string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installIDOrName)
	if err != nil {
		return ui.PrintError(err)
	}

	set, remove, err := labels.ParseArgs(args)
	if err != nil {
		return ui.PrintError(err)
	}

	if len(set) > 0 {
		if _, err := s.api.AddInstallLabels(ctx, installID, set); err != nil {
			return ui.PrintError(fmt.Errorf("unable to add labels: %w", err))
		}
	}
	if len(remove) > 0 {
		if _, err := s.api.RemoveInstallLabels(ctx, installID, remove); err != nil {
			return ui.PrintError(fmt.Errorf("unable to remove labels: %w", err))
		}
	}

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		ui.PrintJSON(map[string]any{"id": install.ID, "labels": install.Labels})
		return nil
	}

	if len(set) > 0 || len(remove) > 0 {
		fmt.Printf("install/%s labeled\n", install.ID)
	}
	printLabels(install.ID, install.Labels)
	return nil
}

func printLabels(id string, lbls map[string]string) {
	if len(lbls) == 0 {
		fmt.Printf("%s: (no labels)\n", id)
		return
	}
	keys := make([]string, 0, len(lbls))
	for k := range lbls {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("%s:\n", id)
	for _, k := range keys {
		fmt.Printf("  %s=%s\n", k, lbls[k])
	}
}

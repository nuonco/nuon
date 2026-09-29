package subscriptiontui

import (
	"context"
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const entityListPageSize = 200

func pickSpecificEntities(ctx context.Context, api API, kind string) ([]string, error) {
	switch kind {
	case "installs":
		return pickInstalls(ctx, api)
	case "components":
		return pickAppScopedEntities(ctx, api, kind, loadComponents)
	case "actions":
		return pickAppScopedEntities(ctx, api, kind, loadActions)
	default:
		return nil, fmt.Errorf("no specific picker for kind %q", kind)
	}
}

func pickInstalls(ctx context.Context, api API) ([]string, error) {
	installs, _, err := api.GetAllInstalls(ctx, &models.GetPaginatedQuery{Limit: entityListPageSize})
	if err != nil {
		return nil, fmt.Errorf("list installs: %w", err)
	}
	if len(installs) == 0 {
		return nil, fmt.Errorf("no installs found in this org — create one with `nuon installs create`, or pick a different match mode")
	}

	options := make([]huh.Option[string], 0, len(installs))
	for _, ins := range installs {
		options = append(options, huh.NewOption(displayName(ins.Name, ins.ID), ins.ID))
	}
	sortOptions(options)

	var selected []string
	form := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Pick installs to scope to").
			Description("Space toggles. Enter advances. Type to filter. Leave empty to drop installs from the match.").
			Options(options...).
			Filterable(true).
			Value(&selected),
	)).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return nil, err
	}
	return selected, nil
}

func pickAppScopedEntities(
	ctx context.Context,
	api API,
	kind string,
	load func(ctx context.Context, api API, appID string) ([]huh.Option[string], error),
) ([]string, error) {
	apps, _, err := api.GetApps(ctx, &models.GetPaginatedQuery{Limit: entityListPageSize})
	if err != nil {
		return nil, fmt.Errorf("list apps: %w", err)
	}
	if len(apps) == 0 {
		return nil, fmt.Errorf("no apps found in this org — create one with `nuon apps create`, or use the labels match mode for cross-app rules")
	}

	var appID string
	if len(apps) == 1 {
		appID = apps[0].ID
		fmt.Printf("Using app %s (%s) — only app in org\n", displayName(apps[0].Name, apps[0].ID), apps[0].ID)
	} else {
		appOpts := make([]huh.Option[string], 0, len(apps))
		for _, a := range apps {
			appOpts = append(appOpts, huh.NewOption(displayName(a.Name, a.ID), a.ID))
		}
		sortOptions(appOpts)

		appForm := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Which app owns the %s you want to pick?", kind)).
				Description("Components and actions are app-scoped. For cross-app selection, cancel and use the labels match mode instead.").
				Options(appOpts...).
				Value(&appID),
		)).WithShowHelp(true)

		if err := appForm.Run(); err != nil {
			return nil, err
		}
	}

	options, err := load(ctx, api, appID)
	if err != nil {
		return nil, err
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("no %s found in app %s — create one first or pick a different match mode", kind, appID)
	}
	sortOptions(options)

	var selected []string
	entityForm := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title(fmt.Sprintf("Pick %s to scope to", kind)).
			Description("Space toggles. Enter advances. Type to filter. Leave empty to drop this kind from the match.").
			Options(options...).
			Filterable(true).
			Value(&selected),
	)).WithShowHelp(true)

	if err := entityForm.Run(); err != nil {
		return nil, err
	}
	return selected, nil
}

func loadComponents(ctx context.Context, api API, appID string) ([]huh.Option[string], error) {
	components, _, err := api.GetAppComponents(ctx, appID, &models.GetPaginatedQuery{Limit: entityListPageSize})
	if err != nil {
		return nil, fmt.Errorf("list components for app %s: %w", appID, err)
	}
	options := make([]huh.Option[string], 0, len(components))
	for _, c := range components {
		options = append(options, huh.NewOption(displayName(c.Name, c.ID), c.ID))
	}
	return options, nil
}

func loadActions(ctx context.Context, api API, appID string) ([]huh.Option[string], error) {
	actions, _, err := api.GetActionWorkflows(ctx, appID, &models.GetPaginatedQuery{Limit: entityListPageSize})
	if err != nil {
		return nil, fmt.Errorf("list actions for app %s: %w", appID, err)
	}
	options := make([]huh.Option[string], 0, len(actions))
	for _, a := range actions {
		options = append(options, huh.NewOption(displayName(a.Name, a.ID), a.ID))
	}
	return options, nil
}

func displayName(name, id string) string {
	if name == "" {
		return id
	}
	return fmt.Sprintf("%s (%s)", name, id)
}

func sortOptions(options []huh.Option[string]) {
	sort.Slice(options, func(i, j int) bool {
		return options[i].Key < options[j].Key
	})
}

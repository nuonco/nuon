// pickers.go houses the data-driven entity pickers that fire after the
// main subscription form (tui.go) when the user selects "specific" for a
// given kind. Components and actions are app-scoped on the server, so
// the picker first asks the user to pick an app, then fetches the
// entities for that app. Installs are org-scoped — no app picker step.
//
// Each picker is its own short huh.Form so the option list can be
// fetched between forms (huh's options are evaluated when the Form is
// constructed, not when the field is rendered, so dynamic option lists
// have to be split across forms).
package subscriptiontui

import (
	"context"
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// entityListPageSize is the upper bound we ask the API for when
// populating a picker. Picker UX assumes the full set fits on one screen
// (huh's MultiSelect is filterable so 200 items is browsable). Above
// this the user should narrow with --subscription-json — flagged in the
// "no entities" / "exceeds page size" copy.
const entityListPageSize = 200

// pickSpecificEntities runs the right picker flow for the given kind:
//   - installs:           org-scoped multi-select
//   - components/actions: app picker (auto-skipped when there's exactly
//     one app in the org), then app-scoped entity multi-select
//
// Returns the selected entity IDs in the order the user picked them.
// An empty result is a legitimate outcome — the caller (Run) treats it
// as "this kind matches nothing", same as if the user had picked skip.
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

// pickInstalls fetches every install in the current org and asks the
// user to multi-select. Installs aren't app-scoped, so there's no app
// picker step here — they enumerate org-wide via GetAllInstalls.
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

// pickAppScopedEntities runs the two-step "pick app, then pick entities
// in that app" flow used for components and actions. The loader callback
// is parameterised so this function works for both kinds without
// duplicating the surrounding huh wiring.
//
// The app picker is auto-skipped when the org has exactly one app — the
// pick is degenerate and forcing the user through a 1-option select
// just adds keystrokes. We surface a one-line confirmation so the user
// knows which app the picker is operating on (the alternative — silent
// auto-pick — is confusing when the org is later expanded with a second
// app and the same flow suddenly grows a step).
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

// displayName formats an entity's display label. Falls back to the bare
// ID when name is empty (older entities, entities created via API
// without a name) so the user can still pick them.
func displayName(name, id string) string {
	if name == "" {
		return id
	}
	return fmt.Sprintf("%s (%s)", name, id)
}

// sortOptions sorts huh options alphabetically by Key (the rendered
// label) so picker output is stable regardless of API ordering. The
// underlying API may return rows in created_at order, which is rarely
// what a human wants when scrolling a list to pick from.
func sortOptions(options []huh.Option[string]) {
	sort.Slice(options, func(i, j int) bool {
		return options[i].Key < options[j].Key
	})
}

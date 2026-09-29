// why: Package subscriptiontui owns the interactive picker for webhook
// subscriptions and is intentionally kept in its own subpackage so the
// non-TUI orgs commands (delete / list / get / api-token / etc.) don't
// transitively import lipgloss + bubbletea.
//
// Importing huh at package scope pulls lipgloss into the binary, and
// lipgloss probes the controlling terminal for capabilities (DEC modes
// 2026 / 2027) the first time it renders. On non-TUI command paths those
// probes are sent but the responses arrive after the program has exited
// and leak to the user's shell as raw escape bytes (e.g. `^[[?2026;2$y`).
// Isolating the import here means only commands that actually invoke
// Run() pull the heavy stack in.
package subscriptiontui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type API interface {
	GetApps(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.AppApp, bool, error)
	GetAppComponents(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppComponent, bool, error)
	GetActionWorkflows(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppActionWorkflow, bool, error)
	GetAllInstalls(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.AppInstall, bool, error)
}

// why: Local mirror of the resource taxonomy from
// services/ctl-api/internal/pkg/interests/types.go. The interests package
// lives behind an internal/ boundary that the CLI can't cross, and the
// SDK doesn't model it (the wire shape is just JSON). Keep these
// constants in lockstep with that file.
//
// The order here matters — the picker renders rows in this order, the
// dashboard does the same.
var (
	resourceKinds = []string{
		"installs",
		"stacks",
		"components",
		"sandboxes",
		"install_configurations",
		"runners",
		"actions",
	}

	resourceOps = map[string][]string{
		"installs":               {"provision", "deprovision", "reprovision", "label_added", "app_branch_changed"},
		"stacks":                 {"version_active"},
		"components":             {"deploy", "teardown"},
		"sandboxes":              {"provision", "reprovision", "deprovision"},
		"install_configurations": {"inputs", "secrets"},
		"runners":                {"provision", "reprovision", "inactive", "unhealthy"},
		"actions":                {"run"},
	}

	driftSupported = map[string]bool{
		"components": true,
		"sandboxes":  true,
	}

	componentHealthSupported = map[string]bool{
		"components": true,
	}

	installDegradedSupported = map[string]bool{
		"installs": true,
	}

	outcomeOptions = []huh.Option[string]{
		huh.NewOption("All (every started + terminal event)", "all"),
		huh.NewOption("Completion (terminal only)", "completion").Selected(true),
		huh.NewOption("Failures (failed + cancelled only)", "failures"),
		huh.NewOption("None (mute lifecycle for this resource)", "none"),
	}

	matchKinds = []string{
		string(labels.TargetKindInstalls),
		string(labels.TargetKindComponents),
		string(labels.TargetKindActions),
	}
)

const (
	matchModeSkip     = "skip"
	matchModeAny      = "any"
	matchModeSpecific = "specific"
	matchModeLabels   = "labels"
)

func matchModeOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Skip — exclude this kind from delivery", matchModeSkip).Selected(true),
		huh.NewOption("Any — every entity of this kind in the org", matchModeAny),
		huh.NewOption("Specific — pick individual entities", matchModeSpecific),
		huh.NewOption("Labels — match by label selector (e.g. env=prod)", matchModeLabels),
	}
}

type resourceState struct {
	enabled         bool
	refine          bool
	ops             []string
	outcome         string
	approval        bool
	drift           bool
	componentHealth bool
	installDegraded bool
}

func newResourceState(kind string) *resourceState {
	return &resourceState{
		outcome:         "completion",
		approval:        true,
		drift:           driftSupported[kind],
		componentHealth: componentHealthSupported[kind],
		installDegraded: installDegradedSupported[kind],
	}
}

type kindMatchState struct {
	mode        string
	ids         []string
	selectorRaw string
	excludeRaw  string
}

// why: Run walks the user through an interactive picker that produces the two
// halves of a webhook SubscriptionPayload. Mirrors the dashboard "Slack
// channel subscription" modal flow:
//
//  1. All events vs per-resource interests
//  2. (per-resource) which kinds to enable, then optionally refine each
//  3. Match: every entity vs scoped to specific entities
//  4. (scoped) per-kind mode select (skip / any / specific / labels);
//     for "specific" components/actions a follow-up flow asks for an app
//     and then multi-selects entities within that app
//
// Phases 1–2 are a single huh.Form so navigating back is meaningful.
// Phase 3 (entity picker) is sequential huh.Forms because the option
// lists are data-driven and must be fetched between forms.
//
// Returns interests as `any` so the orgs service can hand the value
// straight to the SDK without taking on a dependency on this package's
// helpers. A nil match preserves the persisted "every entity in the org"
// semantics — same as omitting the "match" key from --subscription-json.
//
// Caller MUST gate this on cfg.Interactive — non-interactive sessions
// should fall through to the implicit AllEvents default in
// SubscriptionFlags.Resolve.
func Run(ctx context.Context, api API) (any, *labels.SubscriptionMatch, error) {
	allEvents := true
	resources := make(map[string]*resourceState, len(resourceKinds))
	for _, kind := range resourceKinds {
		resources[kind] = newResourceState(kind)
	}

	scoped := false
	kindStates := make(map[string]*kindMatchState, len(matchKinds))
	for _, kind := range matchKinds {
		kindStates[kind] = &kindMatchState{mode: matchModeSkip}
	}

	groups := buildPhase1Groups(&allEvents, resources, &scoped, kindStates)

	if err := huh.NewForm(groups...).WithShowHelp(true).Run(); err != nil {
		return nil, nil, err
	}

	interestsCfg := buildInterests(allEvents, resources)

	if scoped {
		for _, kind := range matchKinds {
			if kindStates[kind].mode != matchModeSpecific {
				continue
			}
			ids, err := pickSpecificEntities(ctx, api, kind)
			if err != nil {
				return nil, nil, fmt.Errorf("pick %s: %w", kind, err)
			}
			kindStates[kind].ids = ids
		}
	}

	match, err := buildMatch(scoped, kindStates)
	if err != nil {
		return nil, nil, err
	}
	return interestsCfg, match, nil
}

func buildPhase1Groups(
	allEvents *bool,
	resources map[string]*resourceState,
	scoped *bool,
	kindStates map[string]*kindMatchState,
) []*huh.Group {
	groups := []*huh.Group{
		huh.NewGroup(
			huh.NewConfirm().
				Title("Send all events?").
				Description("Every workflow, deploy, sandbox, runner, and approval. Turn off to pick which resources and outcomes to receive.").
				Affirmative("Yes — all events").
				Negative("No — pick resources").
				Value(allEvents),
		),
	}

	resourceToggles := make([]huh.Field, 0, len(resourceKinds))
	for _, kind := range resourceKinds {
		st := resources[kind]
		resourceToggles = append(resourceToggles,
			huh.NewConfirm().
				Title(fmt.Sprintf("Enable %s?", kind)).
				Affirmative("Yes").
				Negative("No").
				Value(&st.enabled),
		)
	}
	groups = append(groups, huh.NewGroup(resourceToggles...).
		Title("Which resources?").
		Description("Each enabled resource starts with sensible defaults: completion-only lifecycle, approvals on, drift and health on (where supported). Refine on the next screen if needed.").
		WithHideFunc(func() bool { return *allEvents }))

	for _, kind := range resourceKinds {
		k := kind
		st := resources[k]

		groups = append(groups, huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Refine %s?", k)).
				Description("Defaults: completion outcome, approvals on, drift and health on (if supported), every op.").
				Affirmative("Yes — refine").
				Negative("No — keep defaults").
				Value(&st.refine),
		).WithHideFunc(func() bool { return *allEvents || !st.enabled }))

		fields := []huh.Field{
			huh.NewMultiSelect[string]().
				Title(fmt.Sprintf("%s — ops to include", k)).
				Description("Space toggles. Enter advances. Leave nothing selected to subscribe to every op for this resource.").
				Options(huh.NewOptions(resourceOps[k]...)...).
				Value(&st.ops),

			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s — outcome filter", k)).
				Description(`"completion" is the dashboard default. "none" mutes lifecycle entirely (use with approvals/drift).`).
				Options(outcomeOptions...).
				Value(&st.outcome),

			huh.NewConfirm().
				Title(fmt.Sprintf("%s — approval events?", k)).
				Description("Includes both approval requests and responses (matches the dashboard slack modal).").
				Value(&st.approval),
		}

		if driftSupported[k] {
			fields = append(fields,
				huh.NewConfirm().
					Title(fmt.Sprintf("%s — drift detected events?", k)).
					Description("Fires only when a drift scan finds actual changes.").
					Value(&st.drift),
			)
		}

		if componentHealthSupported[k] {
			fields = append(fields,
				huh.NewConfirm().
					Title(fmt.Sprintf("%s — health change events?", k)).
					Description("Fires when live health crosses into degraded or unhealthy, and again when it recovers. An unknown verdict never notifies.").
					Value(&st.componentHealth),
			)
		}

		if installDegradedSupported[k] {
			fields = append(fields,
				huh.NewConfirm().
					Title(fmt.Sprintf("%s — degraded events?", k)).
					Description("Fires when the install's composite health crosses into degraded or unhealthy, and again when it returns to healthy.").
					Value(&st.installDegraded),
			)
		}

		groups = append(groups, huh.NewGroup(fields...).
			WithHideFunc(func() bool { return *allEvents || !st.enabled || !st.refine }))
	}

	groups = append(groups, huh.NewGroup(
		huh.NewConfirm().
			Title("Scope to specific entities?").
			Description("Defaults to every entity in the org. Turn on to filter by entity kind, IDs, and/or labels.").
			Affirmative("Yes — scope it").
			Negative("No — every entity").
			Value(scoped),
	))

	// why: Per-kind mode select. Each kind gets two groups: a select for
	// the mode, and a hidden-by-default text input for the labels
	// branch (huh v0.8.0 only supports WithHideFunc at the Group
	// level, not per-Field).
	//
	// "Specific" mode defers entity collection to phase 3 (pickers.go)
	// because the option list has to be fetched from the API.
	for _, kind := range matchKinds {
		k := kind
		st := kindStates[k]
		groups = append(groups, huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s — match mode", k)).
				Description("Skip excludes this kind. Any matches every entity. Specific opens a picker. Labels takes a key=value selector.").
				Options(matchModeOptions()...).
				Value(&st.mode),
		).WithHideFunc(func() bool { return !*scoped }))

		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("%s — include label selector", k)).
				Description("key=value pairs, comma-separated (e.g. env=prod,team=*). Leave blank to match everything except the excludes.").
				Value(&st.selectorRaw),
			huh.NewInput().
				Title(fmt.Sprintf("%s — exclude label selector", k)).
				Description("key=value pairs whose match removes an entity (e.g. env=stage). Optional; at least one of include/exclude is required for the Labels mode.").
				Value(&st.excludeRaw),
		).WithHideFunc(func() bool { return !*scoped || st.mode != matchModeLabels }))
	}

	return groups
}

func buildInterests(allEvents bool, resources map[string]*resourceState) any {
	if allEvents {
		return map[string]any{"all_events": true}
	}
	out := map[string]any{}
	for _, kind := range resourceKinds {
		st := resources[kind]
		if !st.enabled {
			continue
		}
		cfg := map[string]any{}
		if len(st.ops) > 0 {
			cfg["ops"] = st.ops
		}
		if st.outcome != "" {
			cfg["outcome"] = st.outcome
		}
		if st.approval {
			cfg["approval_requests"] = true
			cfg["approval_responses"] = true
		}
		if st.drift && driftSupported[kind] {
			cfg["drift_detected"] = true
		}
		if st.componentHealth && componentHealthSupported[kind] {
			cfg["component_health"] = true
		}
		if st.installDegraded && installDegradedSupported[kind] {
			cfg["install_degraded"] = true
		}
		out[kind] = cfg
	}
	return map[string]any{"resources": out}
}

func buildMatch(scoped bool, kindStates map[string]*kindMatchState) (*labels.SubscriptionMatch, error) {
	if !scoped {
		return nil, nil
	}
	m := &labels.SubscriptionMatch{}
	populated := false
	for _, kind := range matchKinds {
		st := kindStates[kind]
		var tm *labels.TargetMatch
		switch st.mode {
		case matchModeSkip:
			continue
		case matchModeAny:
			tm = &labels.TargetMatch{}
		case matchModeSpecific:
			if len(st.ids) == 0 {
				// why: Picker returned no IDs (user unselected everything
				// or there were no entities to pick). Treat as skip
				// rather than poisoning the wire payload with an
				// empty selector that would silently behave like
				// "any" — the user's intent was clearly "specific
				// ones I haven't picked yet".
				continue
			}
			tm = &labels.TargetMatch{IDs: st.ids}
		case matchModeLabels:
			inc := labels.ParseLabelsQuery(st.selectorRaw)
			exc := labels.ParseLabelsQuery(st.excludeRaw)
			if len(inc) == 0 && len(exc) == 0 {
				return nil, fmt.Errorf("%s match mode is labels but no include/exclude key=value pairs were given", kind)
			}
			sel := &labels.Selector{}
			if len(inc) > 0 {
				sel.MatchLabels = inc
			}
			if len(exc) > 0 {
				sel.NotMatchLabels = exc
			}
			tm = &labels.TargetMatch{Selector: sel}
		default:
			return nil, fmt.Errorf("unknown %s match mode %q", kind, st.mode)
		}
		populated = true
		switch labels.TargetKind(kind) {
		case labels.TargetKindInstalls:
			m.Installs = tm
		case labels.TargetKindComponents:
			m.Components = tm
		case labels.TargetKindActions:
			m.Actions = tm
		}
	}
	if !populated {
		return nil, nil
	}
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("invalid match: %w", err)
	}
	return m, nil
}

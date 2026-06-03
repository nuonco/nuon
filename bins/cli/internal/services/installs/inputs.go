package installs

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

// inputDef describes a single declared app input, with its default value.
type inputDef struct {
	name      string
	def       string
	sensitive bool
}

// loadInputDefs fetches the install's app input config and returns the
// declared inputs (name + default) along with the resolved install ID.
func (s *Service) loadInputDefs(ctx context.Context, installID string) (string, []inputDef, error) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return "", nil, err
	}

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return "", nil, err
	}

	cfg, err := s.api.GetAppInputLatestConfig(ctx, install.AppID)
	if err != nil {
		return "", nil, err
	}

	defs := make([]inputDef, 0, len(cfg.Inputs))
	for _, in := range cfg.Inputs {
		defs = append(defs, inputDef{
			name:      in.Name,
			def:       in.Default,
			sensitive: in.Sensitive,
		})
	}
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].name < defs[j].name
	})
	return installID, defs, nil
}

// GetInputs prints the install's current inputs alongside their declared
// defaults.
func (s *Service) GetInputs(ctx context.Context, installID string, asJSON bool) error {
	installID, defs, err := s.loadInputDefs(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	current, err := s.api.GetInstallCurrentInputs(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		ui.PrintJSON(current)
		return nil
	}

	values := redactedValues(current)

	view := ui.NewGetView()
	data := [][]string{{"", "VALUE", "DEFAULT"}}
	for _, d := range defs {
		val, ok := values[d.name]
		if !ok {
			val = "-"
		}
		def := d.def
		if def == "" {
			def = "-"
		}
		data = append(data, []string{styles.TextPrimary.Render(d.name), val, def})
	}
	view.Render(data)
	return nil
}

// SetInputs patches install inputs from a list of key=value pairs. It fetches
// the current inputs first so it can show which values changed, and validates
// that each key refers to a declared input.
func (s *Service) SetInputs(ctx context.Context, installID string, args []string, deployDependents bool, asJSON bool) error {
	updates, err := parseInputArgs(args)
	if err != nil {
		return ui.PrintError(err)
	}

	installID, defs, err := s.loadInputDefs(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	known := make(map[string]inputDef, len(defs))
	for _, d := range defs {
		known[d.name] = d
	}

	var unknown []string
	for k := range updates {
		if _, ok := known[k]; !ok {
			unknown = append(unknown, k)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return ui.PrintError(fmt.Errorf("unknown input(s): %s", strings.Join(unknown, ", ")))
	}

	current, err := s.api.GetInstallCurrentInputs(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	prevValues := currentValues(current)
	prevRedacted := redactedValues(current)

	// The update endpoint expects the full set of inputs, so start from the
	// existing values and merge in only the keys that are being changed.
	merged := make(map[string]string, len(prevValues)+len(updates))
	for k, v := range prevValues {
		merged[k] = v
	}

	// Track which inputs actually changed (new value differs from the
	// original), so we can render an accurate CHANGED column.
	changed := make(map[string]bool, len(updates))
	for k, v := range updates {
		prev, ok := prevValues[k]
		if !ok || prev != v {
			changed[k] = true
		}
		merged[k] = v
	}

	request := &models.ServiceUpdateInstallInputsRequest{
		Inputs:           merged,
		DeployDependents: &deployDependents,
	}
	if config.Debug() {
		ui.PrintJSON(request)
	}

	resp, err := s.api.UpdateInstallInputs(ctx, installID, request)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	if asJSON {
		ui.PrintJSON(resp)
		return nil
	}

	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	view := ui.NewGetView()
	data := [][]string{{"", "PREVIOUS", "NEW", "CHANGED"}}
	for _, k := range keys {
		newVal := updates[k]
		displayNew := newVal
		displayPrev, hadPrev := prevRedacted[k]
		if !hadPrev {
			displayPrev = "-"
		}
		if known[k].sensitive {
			displayNew = "********"
		}
		changedStr := "no"
		if changed[k] {
			// Highlight changes: red for the old value, green for the new.
			displayPrev = styles.TextError.Render(displayPrev)
			displayNew = styles.TextSuccess.Render(displayNew)
			changedStr = styles.TextSuccess.Render("yes")
		}
		data = append(data, []string{styles.TextPrimary.Render(k), displayPrev, displayNew, changedStr})
	}
	view.Render(data)
	return nil
}

// currentValues returns the non-redacted values map from a current-inputs
// response, falling back to an empty map.
func currentValues(in *models.AppInstallInputs) map[string]string {
	if in == nil || in.Values == nil {
		return map[string]string{}
	}
	return in.Values
}

// redactedValues returns the redacted values map, falling back to an empty map.
func redactedValues(in *models.AppInstallInputs) map[string]string {
	if in == nil || in.RedactedValues == nil {
		return map[string]string{}
	}
	return in.RedactedValues
}

// parseInputArgs parses a list of key=value pairs into a map. It errors on
// malformed entries.
func parseInputArgs(args []string) (map[string]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no inputs provided; expected key=value pairs")
	}
	out := make(map[string]string, len(args))
	for _, arg := range args {
		k, v, ok := strings.Cut(arg, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("invalid input %q; expected key=value", arg)
		}
		out[k] = v
	}
	return out, nil
}

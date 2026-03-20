package installs

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) SandboxOutputs(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	runs, _, err := s.api.GetInstallSandboxRuns(ctx, installID, &models.GetPaginatedQuery{Limit: 50})
	if err != nil {
		return ui.PrintError(err)
	}

	// Find latest successful/active run with outputs.
	var outputs map[string]any
	for _, run := range runs {
		if (run.Status == "active" || run.Status == "succeeded") && run.Outputs != nil {
			outputs = run.Outputs
			break
		}
	}
	if outputs == nil && len(runs) > 0 && runs[0].Outputs != nil {
		outputs = runs[0].Outputs
	}
	if outputs == nil {
		view := ui.NewGetView()
		view.Print("No sandbox outputs available for this install.")
		return nil
	}

	if asJSON {
		ui.PrintJSON(outputs)
		return nil
	}

	// Flatten nested map into dot-notation key/value pairs for table display.
	flat := make(map[string]string)
	flattenMap("", outputs, flat)

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	data := [][]string{{"KEY", "VALUE"}}
	for _, k := range keys {
		data = append(data, []string{k, flat[k]})
	}

	view := ui.NewListView()
	view.Render(data)
	return nil
}

// flattenMap recursively flattens a nested map into dot-notation keys.
func flattenMap(prefix string, m map[string]any, out map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			flattenMap(key, val, out)
		case []any:
			parts := make([]string, len(val))
			for i, item := range val {
				parts[i] = fmt.Sprintf("%v", item)
			}
			out[key] = strings.Join(parts, ", ")
		default:
			out[key] = fmt.Sprintf("%v", v)
		}
	}
}

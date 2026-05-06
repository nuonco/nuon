package installs

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/services/labels"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, offset, limit int, labelFilters []string, asJSON bool) error {
	view := ui.NewListView()

	filter, _, err := labels.ParseArgs(labelFilters)
	if err != nil {
		return view.Error(fmt.Errorf("invalid --labels value: %w", err))
	}

	var (
		installs []*models.AppInstall
		hasMore  bool
	)

	if appID != "" {
		resolvedAppID, lerr := lookup.AppID(ctx, s.api, appID)
		if lerr != nil {
			return ui.PrintError(lerr)
		}
		installs, hasMore, err = s.listAppInstalls(ctx, resolvedAppID, offset, limit)
	} else {
		installs, hasMore, err = s.listInstalls(ctx, offset, limit)
	}
	if err != nil {
		return view.Error(err)
	}

	if len(filter) > 0 {
		installs = filterInstallsByLabels(installs, filter)
	}

	if asJSON {
		ui.PrintJSON(installs)
		return nil
	}

	data := [][]string{
		{
			"NAME",
			"ID",
			"SANDBOX",
			"RUNNER",
			"COMPONENTS",
			"LABELS",
			"CREATED AT",
		},
	}
	curID := s.cfg.GetString("org_id")
	for _, install := range installs {
		if curID != "" {
			if install.ID == curID {
				install.Name = "*" + install.Name
			} else {
				install.Name = " " + install.Name
			}
		}
		data = append(data, []string{
			install.Name,
			install.ID,
			install.SandboxStatus,
			install.RunnerStatus,
			install.CompositeComponentStatus,
			formatLabels(install.Labels),
			install.CreatedAt,
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listInstalls(ctx context.Context, offset, limit int) ([]*models.AppInstall, bool, error) {
	installs, hasMore, err := s.api.GetAllInstalls(ctx, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return installs, hasMore, nil
}

func (s *Service) listAppInstalls(ctx context.Context, appID string, offset, limit int) ([]*models.AppInstall, bool, error) {
	cmps, hasMore, err := s.api.GetAppInstalls(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return cmps, hasMore, nil
}

// filterInstallsByLabels keeps only installs whose labels contain every
// key=value pair in filter (AND semantics).
func filterInstallsByLabels(installs []*models.AppInstall, filter map[string]string) []*models.AppInstall {
	out := make([]*models.AppInstall, 0, len(installs))
	for _, i := range installs {
		if installMatchesLabels(i, filter) {
			out = append(out, i)
		}
	}
	return out
}

func installMatchesLabels(install *models.AppInstall, filter map[string]string) bool {
	for k, v := range filter {
		got, ok := install.Labels[k]
		if !ok || got != v {
			return false
		}
	}
	return true
}

func formatLabels(lbls map[string]string) string {
	if len(lbls) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(lbls))
	for k := range lbls {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, lbls[k]))
	}
	return strings.Join(parts, ",")
}

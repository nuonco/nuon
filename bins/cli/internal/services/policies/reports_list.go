package policies

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

func (s *Service) ListReports(ctx context.Context, appID, installID, ownerType, status string, offset, limit int, asJSON bool) error {
	var err error

	if appID != "" {
		appID, err = lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}
	}

	if installID != "" {
		installID, err = lookup.InstallID(ctx, s.api, installID)
		if err != nil {
			return ui.PrintError(err)
		}
	}

	view := ui.NewListView()

	reports, err := s.api.GetPolicyReports(ctx, &nuon.PolicyReportsQuery{
		AppID:     appID,
		InstallID: installID,
		OwnerType: ownerType,
		Status:    status,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return view.Error(errors.Wrap(err, "failed to fetch policy reports"))
	}

	if asJSON {
		ui.PrintJSON(reports)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"OWNER TYPE",
			"STATUS",
			"DENY",
			"WARN",
			"PASS",
			"CREATED AT",
		},
	}
	for _, report := range reports {
		status := ""
		if report.Status != nil {
			status = string(report.Status.Status)
		}
		data = append(data, []string{
			report.ID,
			report.OwnerType,
			status,
			fmt.Sprintf("%d", report.DenyCount),
			fmt.Sprintf("%d", report.WarnCount),
			fmt.Sprintf("%d", report.PassCount),
			report.CreatedAt,
		})
	}

	hasMore := limit > 0 && len(reports) >= limit
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

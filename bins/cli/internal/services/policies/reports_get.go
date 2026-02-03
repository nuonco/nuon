package policies

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) GetReport(ctx context.Context, reportID string, asJSON bool) error {
	view := ui.NewGetView()

	report, err := s.api.GetPolicyReport(ctx, reportID)
	if err != nil {
		return view.Error(errors.Wrap(err, "failed to fetch policy report"))
	}

	if asJSON {
		ui.PrintJSON(report)
		return nil
	}

	status := ""
	if report.Status != nil {
		status = string(report.Status.Status)
	}

	data := [][]string{
		{"id", report.ID},
		{"owner id", report.OwnerID},
		{"owner type", report.OwnerType},
		{"app id", report.AppID},
		{"install id", report.InstallID},
		{"status", status},
		{"deny count", fmt.Sprintf("%d", report.DenyCount)},
		{"warn count", fmt.Sprintf("%d", report.WarnCount)},
		{"pass count", fmt.Sprintf("%d", report.PassCount)},
		{"evaluated at", report.EvaluatedAt},
		{"created at", report.CreatedAt},
	}

	view.Render(data)

	if len(report.Violations) > 0 {
		fmt.Println("\nViolations:")
		violationData := [][]string{
			{"POLICY ID", "INPUT INDEX", "SEVERITY", "MESSAGE"},
		}
		for _, v := range report.Violations {
			violationData = append(violationData, []string{
				v.PolicyID,
				fmt.Sprintf("%d", v.InputIndex),
				v.Severity,
				v.Message,
			})
		}
		violationView := ui.NewListView()
		violationView.RenderPaging(violationData, 0, 0, false)
	}

	return nil
}

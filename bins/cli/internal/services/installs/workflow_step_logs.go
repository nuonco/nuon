package installs

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/logs"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// stepLogFollowWait is how long --tail waits after history is drained and the stream is still open.
var stepLogFollowWait = 5 * time.Second

type WorkflowStepLogsOptions struct {
	Follow    bool
	Raw       bool
	Browser   bool
	Limit     int
	Filter    string
	Severity  []string
	Service   []string
	SortOrder string
}

func (s *Service) WorkflowStepLogs(ctx context.Context, installID, workflowID, stepID string, asJSON bool, opts WorkflowStepLogsOptions) error {
	view := ui.NewListView()

	// If stepID is not provided, use the last processed step (not the one awaiting action)
	if stepID == "" {
		var err error
		stepID, err = s.getLastProcessedStepID(ctx, workflowID)
		if err != nil {
			return view.Error(err)
		}
	}

	if opts.Browser {
		workflow, err := s.api.GetWorkflow(ctx, workflowID)
		if err != nil {
			return view.Error(err)
		}

		cfg, err := s.api.GetCLIConfig(ctx)
		if err != nil {
			return view.Error(err)
		}

		url := fmt.Sprintf("%s/%s/installs/%s/workflows/%s?target=%s", cfg.DashboardURL, s.cfg.OrgID, workflow.OwnerID, workflowID, stepID)
		ui.Println(url)
		browser.OpenURL(url)
		return nil
	}

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	step, err := s.api.GetWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		return view.Error(err)
	}

	if step.StepTargetID == "" {
		return view.Error(fmt.Errorf("step %s does not have a target", stepID))
	}

	var logStreamID string
	var deployID string
	switch step.StepTargetType {
	case "install_deploys":
		deployID = step.StepTargetID
		deploy, err := s.api.GetInstallDeploy(ctx, installID, deployID)
		if err != nil {
			return view.Error(err)
		}
		if deploy.LogStream != nil {
			logStreamID = deploy.LogStream.ID
		}
	case "install_action_workflow_runs":
		run, err := s.api.GetInstallActionWorkflowRun(ctx, installID, step.StepTargetID)
		if err != nil {
			return view.Error(err)
		}
		if run.LogStream != nil {
			logStreamID = run.LogStream.ID
		}
	case "install_sandbox_runs":
		run, err := s.api.GetInstallSandboxRun(ctx, installID, step.StepTargetID)
		if err != nil {
			return view.Error(err)
		}
		if run.LogStream != nil {
			logStreamID = run.LogStream.ID
		}
	default:
		return view.Error(fmt.Errorf("unsupported step target type: %s", step.StepTargetType))
	}

	if logStreamID == "" {
		return view.Error(fmt.Errorf("no log stream found for step %s", stepID))
	}

	if asJSON {
		logRecords, err := s.readAllStepLogs(ctx, logStreamID)
		if err != nil {
			return view.Error(err)
		}
		ui.PrintJSON(logRecords)
		return nil
	}

	if opts.Raw || opts.Follow || opts.Limit > 0 || opts.Filter != "" || len(opts.Severity) > 0 || len(opts.Service) > 0 || opts.SortOrder != "" {
		return s.streamStepLogs(ctx, logStreamID, opts)
	}

	logs.LogStreamApp(ctx, s.cfg, s.api, installID, deployID, logStreamID)
	return nil
}

func (s *Service) streamStepLogs(ctx context.Context, logStreamID string, opts WorkflowStepLogsOptions) error {
	// Build severity and service filter sets for O(1) lookup
	severitySet := make(map[string]bool, len(opts.Severity))
	for _, sev := range opts.Severity {
		severitySet[strings.ToLower(sev)] = true
	}
	serviceSet := make(map[string]bool, len(opts.Service))
	for _, svc := range opts.Service {
		serviceSet[strings.ToLower(svc)] = true
	}

	// For sort support, we need to collect all records first then sort
	needsSort := opts.SortOrder == "asc" || opts.SortOrder == "desc"
	shouldCollect := needsSort && !opts.Follow

	var collected []*models.AppOtelLogRecord

	offset := ""
	printed := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		logRecords, nextOffset, err := s.api.LogStreamReadLogsWithNextOffset(ctx, logStreamID, offset, "", nil)
		if err != nil {
			return ui.PrintError(err)
		}

		for _, rec := range logRecords {
			if !matchesLogFilters(rec, opts.Filter, severitySet, serviceSet) {
				continue
			}

			if shouldCollect {
				collected = append(collected, rec)
			} else {
				ui.Printf("[%s] %s %s: %s\n", rec.Timestamp, rec.SeverityText, rec.ServiceName, rec.Body)
				printed++
				if opts.Limit > 0 && printed >= opts.Limit {
					return nil
				}
			}
		}

		if nextOffset != "" && nextOffset != offset {
			offset = nextOffset
			continue
		}

		if !opts.Follow {
			break
		}

		// A short page returns an empty next header. Follow from the last
		// record so the next poll does not reprint this page.
		if cursor := lastLogCursor(logRecords); cursor != "" {
			offset = cursor
		}

		logStream, err := s.api.GetLogStream(ctx, logStreamID)
		if err != nil || !logStream.Open {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(stepLogFollowWait):
		}
	}

	if shouldCollect {
		sortLogRecords(collected, opts.SortOrder)
		for _, rec := range collected {
			ui.Printf("[%s] %s %s: %s\n", rec.Timestamp, rec.SeverityText, rec.ServiceName, rec.Body)
			printed++
			if opts.Limit > 0 && printed >= opts.Limit {
				return nil
			}
		}
	}

	return nil
}

func (s *Service) readAllStepLogs(ctx context.Context, logStreamID string) ([]*models.AppOtelLogRecord, error) {
	var all []*models.AppOtelLogRecord
	offset := ""
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		records, next, err := s.api.LogStreamReadLogsWithNextOffset(ctx, logStreamID, offset, "", nil)
		if err != nil {
			return nil, err
		}
		all = append(all, records...)
		if next == "" || next == offset {
			if all == nil {
				all = []*models.AppOtelLogRecord{}
			}
			return all, nil
		}
		offset = next
	}
}

func lastLogCursor(records []*models.AppOtelLogRecord) string {
	if len(records) == 0 {
		return ""
	}
	return unixNanoCursor(records[len(records)-1].Timestamp)
}

func unixNanoCursor(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		parsed, err := time.Parse(layout, timestamp)
		if err == nil {
			return strconv.FormatInt(parsed.UnixNano(), 10)
		}
	}
	return ""
}

func matchesLogFilters(rec *models.AppOtelLogRecord, filter string, severitySet, serviceSet map[string]bool) bool {
	if len(severitySet) > 0 && !severitySet[strings.ToLower(rec.SeverityText)] {
		return false
	}
	if len(serviceSet) > 0 && !serviceSet[strings.ToLower(rec.ServiceName)] {
		return false
	}
	if filter != "" && !strings.Contains(strings.ToLower(rec.Body), strings.ToLower(filter)) {
		return false
	}
	return true
}

func sortLogRecords(records []*models.AppOtelLogRecord, order string) {
	sort.Slice(records, func(i, j int) bool {
		if order == "asc" {
			return records[i].Timestamp < records[j].Timestamp
		}
		return records[i].Timestamp > records[j].Timestamp
	})
}

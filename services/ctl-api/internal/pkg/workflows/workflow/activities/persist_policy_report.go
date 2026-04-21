package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/helpers"
)

type PersistPolicyReportRequest struct {
	OrgID                          string            `json:"org_id" validate:"required"`
	AppID                          string            `json:"app_id" validate:"required"`
	InstallID                      *string           `json:"install_id"`
	InstallSandboxID               *string           `json:"install_sandbox_id"`
	ComponentID                    *string           `json:"component_id"`
	ComponentBuildID               *string           `json:"component_build_id"`
	WorkflowStepPolicyValidationID *string           `json:"workflow_step_policy_validation_id"`
	RunnerJobID                    *string           `json:"runner_job_id"`
	OwnerID                        string            `json:"owner_id" validate:"required"`
	OwnerType                      string            `json:"owner_type" validate:"required"`
	Violations                     []PolicyViolation `json:"violations"`
	PolicyIDs                      []string          `json:"policy_ids"`
	PolicyInputCounts              map[string]int    `json:"policy_input_counts"`
	DenyCount                      int               `json:"deny_count"`
	WarnCount                      int               `json:"warn_count"`
	PassCount                      int               `json:"pass_count"`

	// Human-readable names for display in reports
	OrgName       string `json:"org_name"`
	AppName       string `json:"app_name"`
	InstallName   string `json:"install_name"`
	ComponentName string `json:"component_name"`
}

type PersistPolicyReportResult struct {
	ReportID        string   `json:"report_id" temporaljson:"report_id,omitempty"`
	DenyCount       int      `json:"deny_count" temporaljson:"deny_count,omitempty"`
	WarnCount       int      `json:"warn_count" temporaljson:"warn_count,omitempty"`
	PassCount       int      `json:"pass_count" temporaljson:"pass_count,omitempty"`
	PassedPolicyIDs []string `json:"passed_policy_ids" temporaljson:"passed_policy_ids,omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 3
// @schedule-to-close-timeout 2m
// @start-to-close-timeout 1m30s
func (a *Activities) PersistPolicyReport(ctx context.Context, req *PersistPolicyReportRequest) (*PersistPolicyReportResult, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(
		zap.String("org_id", req.OrgID),
		zap.String("app_id", req.AppID),
		zap.String("owner_id", req.OwnerID),
		zap.String("owner_type", req.OwnerType),
		zap.Int("violations_count", len(req.Violations)),
	)
	if req.WorkflowStepPolicyValidationID != nil {
		l = l.With(zap.String("validation_id", *req.WorkflowStepPolicyValidationID))
	}

	l.Info("persisting policy report")

	denyCount := req.DenyCount
	warnCount := req.WarnCount
	passCount := req.PassCount
	if denyCount == 0 && warnCount == 0 && passCount == 0 && len(req.Violations) > 0 {
		for _, v := range req.Violations {
			switch v.Severity {
			case "deny":
				denyCount++
			case "warn":
				warnCount++
			}
		}
	}

	policyResults := buildPolicyResults(req.PolicyIDs, req.PolicyInputCounts, req.Violations)
	if passCount == 0 && denyCount == 0 && warnCount == 0 {
		for _, result := range policyResults {
			denyCount += result.DenyCount
			warnCount += result.WarnCount
			passCount += result.PassCount
		}
	}

	violations := req.Violations

	policies := helpers.ToAppResultsFromInternal(toInternalResults(policyResults))

	inputs := buildPolicyInputRefs(req)
	appInputs := helpers.ToAppInputRefsFromInternal(toInternalInputRefs(inputs))

	report := &app.PolicyReport{
		OrgID:                          req.OrgID,
		AppID:                          req.AppID,
		InstallID:                      req.InstallID,
		ComponentID:                    req.ComponentID,
		WorkflowStepPolicyValidationID: req.WorkflowStepPolicyValidationID,
		RunnerJobID:                    req.RunnerJobID,
		OwnerID:                        req.OwnerID,
		OwnerType:                      app.PolicyReportOwnerType(req.OwnerType),
		EvaluatedAt:                    time.Now().UTC(),
		Violations:                     violations,
		PolicyIDs:                      req.PolicyIDs,
		Policies:                       policies,
		Inputs:                         appInputs,
		DenyCount:                      denyCount,
		WarnCount:                      warnCount,
		PassCount:                      passCount,
		Status:                         buildPolicyReportStatus(ctx, denyCount, warnCount, passCount),
		// Human-readable names for display in reports
		OrgName:       req.OrgName,
		AppName:       req.AppName,
		InstallName:   stringPtrToPtr(req.InstallName),
		ComponentName: stringPtrToPtr(req.ComponentName),
	}

	if err := a.db.WithContext(ctx).Create(report).Error; err != nil {
		l.Error("failed to persist policy report", zap.Error(err))
		return nil, errors.Wrap(err, "failed to persist policy report")
	}

	l.Info("policy report persisted successfully",
		zap.String("report_id", report.ID),
		zap.Int("deny_count", denyCount),
		zap.Int("warn_count", warnCount),
		zap.Int("pass_count", passCount),
	)

	// Write analytics events to ClickHouse (non-blocking — failures don't affect the activity)
	a.persistPolicyAnalyticsEvents(ctx, l, report, policyResults)

	passedPolicyIDs := make([]string, 0)
	for _, result := range policyResults {
		if result.Status == "pass" {
			passedPolicyIDs = append(passedPolicyIDs, result.PolicyID)
		}
	}

	return &PersistPolicyReportResult{
		ReportID:        report.ID,
		DenyCount:       denyCount,
		WarnCount:       warnCount,
		PassCount:       passCount,
		PassedPolicyIDs: passedPolicyIDs,
	}, nil
}

type policyResult struct {
	PolicyID   string
	Status     string
	DenyCount  int
	WarnCount  int
	PassCount  int
	InputCount int
}

func buildPolicyResults(policyIDs []string, policyInputCounts map[string]int, violations []PolicyViolation) []policyResult {
	if len(policyIDs) == 0 {
		policyIDs = make([]string, 0)
		seen := make(map[string]struct{})
		for _, violation := range violations {
			if violation.PolicyID == "" {
				continue
			}
			if _, exists := seen[violation.PolicyID]; exists {
				continue
			}
			seen[violation.PolicyID] = struct{}{}
			policyIDs = append(policyIDs, violation.PolicyID)
		}
	}

	if len(policyIDs) == 0 {
		return []policyResult{}
	}

	results := make([]policyResult, 0, len(policyIDs))
	for _, policyID := range policyIDs {
		results = append(results, policyResult{PolicyID: policyID})
	}

	resultIndex := make(map[string]int, len(policyIDs))
	for i, policyID := range policyIDs {
		resultIndex[policyID] = i
	}

	for _, violation := range violations {
		idx, ok := resultIndex[violation.PolicyID]
		if !ok {
			continue
		}
		results[idx].InputCount = maxInt(results[idx].InputCount, violation.InputIndex+1)
		switch violation.Severity {
		case "deny":
			results[idx].DenyCount++
		case "warn":
			results[idx].WarnCount++
		}
	}

	for i := range results {
		result := &results[i]
		inputCount := result.InputCount
		if policyInputCounts != nil {
			if total, ok := policyInputCounts[result.PolicyID]; ok {
				inputCount = maxInt(inputCount, total)
			}
		}
		result.InputCount = inputCount
		if result.DenyCount > 0 {
			result.Status = "deny"
		} else if result.WarnCount > 0 {
			result.Status = "warn"
		} else {
			result.Status = "pass"
			result.PassCount = maxInt(1, result.InputCount)
		}
	}

	return results
}

type policyInputRef struct {
	ID   string
	Type string
	Name string
}

func buildPolicyInputRefs(req *PersistPolicyReportRequest) []policyInputRef {
	refs := make([]policyInputRef, 0, 3)
	addRef := func(id, refType, name string) {
		if id == "" {
			return
		}
		refs = append(refs, policyInputRef{
			ID:   id,
			Type: refType,
			Name: name,
		})
	}

	switch req.OwnerType {
	case string(app.PolicyReportOwnerTypeComponentBuild):
		addRef(req.OwnerID, "component_build", "")
		if req.ComponentID != nil {
			addRef(*req.ComponentID, "component", req.ComponentName)
		}
		return refs
	case string(app.PolicyReportOwnerTypeInstallSandboxRun):
		addRef(req.OwnerID, "sandbox_run", "")
		if req.InstallSandboxID != nil {
			addRef(*req.InstallSandboxID, "sandbox", "")
		}
		return refs
	case string(app.PolicyReportOwnerTypeInstallDeploy):
		addRef(req.OwnerID, "component_deploy", "")
		if req.ComponentBuildID != nil {
			addRef(*req.ComponentBuildID, "component_build", "")
		}
		if req.ComponentID != nil {
			addRef(*req.ComponentID, "component", req.ComponentName)
		}
		return refs
	default:
		return refs
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func buildPolicyReportStatus(ctx context.Context, denyCount, warnCount, passCount int) app.CompositeStatus {
	status := app.StatusSuccess
	statusDescription := "policy checks passed"

	if denyCount > 0 {
		status = app.StatusError
		statusDescription = "policy checks failed"
	} else if warnCount > 0 {
		status = app.StatusWarning
		statusDescription = "policy warnings detected"
	}

	composite := app.NewCompositeStatus(ctx, status)
	composite.StatusHumanDescription = statusDescription
	composite.Metadata = map[string]any{
		"deny_count": denyCount,
		"warn_count": warnCount,
		"pass_count": passCount,
	}
	return composite
}

func stringPtrToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (a *Activities) persistPolicyAnalyticsEvents(ctx context.Context, l *zap.Logger, report *app.PolicyReport, results []policyResult) {
	if a.chDB == nil || len(results) == 0 {
		return
	}

	events := buildPolicyReportEvents(report, results)
	if err := a.chDB.WithContext(ctx).Create(&events).Error; err != nil {
		l.Warn("failed to write policy analytics to clickhouse", zap.Error(err))
	}
}

func buildPolicyReportEvents(report *app.PolicyReport, results []policyResult) []app.PolicyReportEvent {
	events := make([]app.PolicyReportEvent, 0, len(results))
	for _, pr := range results {
		events = append(events, app.PolicyReportEvent{
			ReportID:    report.ID,
			OrgID:       report.OrgID,
			AppID:       report.AppID,
			InstallID:   generics.FromPtrStr(report.InstallID),
			ComponentID: generics.FromPtrStr(report.ComponentID),
			PolicyID:    pr.PolicyID,
			OwnerType:   string(report.OwnerType),
			EvaluatedAt: report.EvaluatedAt,
			Outcome:     pr.Status,
		})
	}
	return events
}

func toInternalResults(results []policyResult) []helpers.PolicyResultInternal {
	internal := make([]helpers.PolicyResultInternal, len(results))
	for i, r := range results {
		internal[i] = helpers.PolicyResultInternal{
			PolicyID:   r.PolicyID,
			Status:     r.Status,
			DenyCount:  r.DenyCount,
			WarnCount:  r.WarnCount,
			PassCount:  r.PassCount,
			InputCount: r.InputCount,
		}
	}
	return internal
}

func toInternalInputRefs(inputs []policyInputRef) []helpers.PolicyInputRefInternal {
	internal := make([]helpers.PolicyInputRefInternal, len(inputs))
	for i, inp := range inputs {
		internal[i] = helpers.PolicyInputRefInternal{
			ID:   inp.ID,
			Type: inp.Type,
			Name: inp.Name,
		}
	}
	return internal
}

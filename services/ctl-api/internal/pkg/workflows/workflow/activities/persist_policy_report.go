package activities

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// SARIF (Static Analysis Results Interchange Format) types for policy evaluation reports
// Based on SARIF 2.1.0 specification: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html

const (
	SARIFVersion   = "2.1.0"
	SARIFSchemaURI = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"
)

type SARIFReport struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

type SARIFTool struct {
	Driver SARIFToolDriver `json:"driver"`
}

type SARIFToolDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version,omitempty"`
	InformationURI string      `json:"informationUri,omitempty"`
	Rules          []SARIFRule `json:"rules,omitempty"`
}

type SARIFRule struct {
	ID               string       `json:"id"`
	ShortDescription SARIFMessage `json:"shortDescription,omitempty"`
}

type SARIFResult struct {
	RuleID  string       `json:"ruleId"`
	Level   string       `json:"level"` // "error", "warning", "note"
	Message SARIFMessage `json:"message"`
}

type SARIFMessage struct {
	Text string `json:"text"`
}

// OPA raw report format
type OPAReport struct {
	EvaluatedAt time.Time         `json:"evaluated_at"`
	Violations  []PolicyViolation `json:"violations"`
	PolicyIDs   []string          `json:"policy_ids"`
	Policies    []PolicyResult    `json:"policies"`
	Inputs      []PolicyInputRef  `json:"inputs,omitempty"`
	DenyCount   int               `json:"deny_count"`
	WarnCount   int               `json:"warn_count"`
	PassCount   int               `json:"pass_count"`
}

type PolicyInputRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type PolicyResult struct {
	PolicyID   string `json:"policy_id"`
	Status     string `json:"status"`
	DenyCount  int    `json:"deny_count"`
	WarnCount  int    `json:"warn_count"`
	PassCount  int    `json:"pass_count"`
	InputCount int    `json:"input_count"`
}

type PersistPolicyReportRequest struct {
	OrgID                          string            `json:"org_id" validate:"required"`
	AppID                          string            `json:"app_id" validate:"required"`
	InstallID                      *string           `json:"install_id"`
	InstallSandboxID               *string           `json:"install_sandbox_id"`
	ComponentID                    *string           `json:"component_id"`
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
}

type PersistPolicyReportResult struct {
	OPAReportID   string `json:"opa_report_id" temporaljson:"opa_report_id,omitempty"`
	SARIFReportID string `json:"sarif_report_id" temporaljson:"sarif_report_id,omitempty"`
	DenyCount     int    `json:"deny_count" temporaljson:"deny_count,omitempty"`
	WarnCount     int    `json:"warn_count" temporaljson:"warn_count,omitempty"`
	PassCount     int    `json:"pass_count" temporaljson:"pass_count,omitempty"`
}

// @temporal-gen activity
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

	l.Info("persisting policy reports")

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

	// Create OPA report
	opaReport := OPAReport{
		EvaluatedAt: time.Now().UTC(),
		Violations:  req.Violations,
		PolicyIDs:   req.PolicyIDs,
		Policies:    policyResults,
		Inputs:      buildPolicyInputRefs(req),
		DenyCount:   denyCount,
		WarnCount:   warnCount,
		PassCount:   passCount,
	}

	opaContent, err := json.Marshal(opaReport)
	if err != nil {
		l.Error("failed to marshal OPA report", zap.Error(err))
		return nil, errors.Wrap(err, "failed to marshal OPA report")
	}

	// Create SARIF report
	sarifReport := a.buildSARIFReport(req.Violations)
	sarifContent, err := json.Marshal(sarifReport)
	if err != nil {
		l.Error("failed to marshal SARIF report", zap.Error(err))
		return nil, errors.Wrap(err, "failed to marshal SARIF report")
	}

	// Persist OPA report
	opaReportRecord := &app.PolicyReport{
		OrgID:                          req.OrgID,
		AppID:                          req.AppID,
		InstallID:                      req.InstallID,
		ComponentID:                    req.ComponentID,
		WorkflowStepPolicyValidationID: req.WorkflowStepPolicyValidationID,
		RunnerJobID:                    req.RunnerJobID,
		OwnerID:                        req.OwnerID,
		OwnerType:                      app.PolicyReportOwnerType(req.OwnerType),
		Format:                         app.PolicyReportFormatOPA,
		ContentVersion:                 "1.0.0",
		Content:                        opaContent,
		DenyCount:                      denyCount,
		WarnCount:                      warnCount,
		PassCount:                      passCount,
		Status:                         buildPolicyReportStatus(ctx, denyCount, warnCount, passCount),
	}

	if err := a.db.WithContext(ctx).Create(opaReportRecord).Error; err != nil {
		l.Error("failed to persist OPA report", zap.Error(err))
		return nil, errors.Wrap(err, "failed to persist OPA report")
	}

	l.Debug("persisted OPA report", zap.String("report_id", opaReportRecord.ID))

	// Persist SARIF report
	sarifReportRecord := &app.PolicyReport{
		OrgID:                          req.OrgID,
		AppID:                          req.AppID,
		InstallID:                      req.InstallID,
		ComponentID:                    req.ComponentID,
		WorkflowStepPolicyValidationID: req.WorkflowStepPolicyValidationID,
		RunnerJobID:                    req.RunnerJobID,
		OwnerID:                        req.OwnerID,
		OwnerType:                      app.PolicyReportOwnerType(req.OwnerType),
		Format:                         app.PolicyReportFormatSARIF,
		ContentVersion:                 SARIFVersion,
		Content:                        sarifContent,
		DenyCount:                      denyCount,
		WarnCount:                      warnCount,
		PassCount:                      passCount,
		Status:                         buildPolicyReportStatus(ctx, denyCount, warnCount, passCount),
	}

	if err := a.db.WithContext(ctx).Create(sarifReportRecord).Error; err != nil {
		l.Error("failed to persist SARIF report", zap.Error(err))
		return nil, errors.Wrap(err, "failed to persist SARIF report")
	}

	l.Debug("persisted SARIF report", zap.String("report_id", sarifReportRecord.ID))

	l.Info("policy reports persisted successfully",
		zap.String("opa_report_id", opaReportRecord.ID),
		zap.String("sarif_report_id", sarifReportRecord.ID),
		zap.Int("deny_count", denyCount),
		zap.Int("warn_count", warnCount),
	)

	return &PersistPolicyReportResult{
		OPAReportID:   opaReportRecord.ID,
		SARIFReportID: sarifReportRecord.ID,
		DenyCount:     denyCount,
		WarnCount:     warnCount,
		PassCount:     passCount,
	}, nil
}

func buildPolicyResults(policyIDs []string, policyInputCounts map[string]int, violations []PolicyViolation) []PolicyResult {
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
		return []PolicyResult{}
	}

	results := make([]PolicyResult, 0, len(policyIDs))
	for _, policyID := range policyIDs {
		results = append(results, PolicyResult{PolicyID: policyID})
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

func buildPolicyInputRefs(req *PersistPolicyReportRequest) []PolicyInputRef {
	refs := make([]PolicyInputRef, 0, 2)
	addRef := func(id, refType string) {
		if id == "" {
			return
		}
		refs = append(refs, PolicyInputRef{
			ID:   id,
			Type: refType,
		})
	}

	switch req.OwnerType {
	case string(app.PolicyReportOwnerTypeComponentBuild):
		addRef(req.OwnerID, "component_build")
		if req.ComponentID != nil {
			addRef(*req.ComponentID, "component")
		}
		return refs
	case string(app.PolicyReportOwnerTypeInstallSandboxRun):
		addRef(req.OwnerID, "sandbox_run")
		if req.InstallSandboxID != nil {
			addRef(*req.InstallSandboxID, "sandbox")
		}
		return refs
	case string(app.PolicyReportOwnerTypeInstallDeploy):
		addRef(req.OwnerID, "component_deploy")
		if req.ComponentID != nil {
			addRef(*req.ComponentID, "component")
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
		status = app.StatusInProgress
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

func (a *Activities) buildSARIFReport(violations []PolicyViolation) SARIFReport {
	// Collect unique policy IDs for rules
	policyIDs := make(map[string]bool)
	for _, v := range violations {
		policyIDs[v.PolicyID] = true
	}

	rules := make([]SARIFRule, 0, len(policyIDs))
	for policyID := range policyIDs {
		rules = append(rules, SARIFRule{
			ID: policyID,
			ShortDescription: SARIFMessage{
				Text: "Policy: " + policyID,
			},
		})
	}

	// Convert violations to SARIF results
	results := make([]SARIFResult, len(violations))
	for i, v := range violations {
		level := "warning"
		if v.Severity == "deny" {
			level = "error"
		}

		results[i] = SARIFResult{
			RuleID: v.PolicyID,
			Level:  level,
			Message: SARIFMessage{
				Text: v.Message,
			},
		}
	}

	return SARIFReport{
		Version: SARIFVersion,
		Schema:  SARIFSchemaURI,
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFToolDriver{
						Name:           "nuon-policy",
						Version:        "1.0.0",
						InformationURI: "https://nuon.co",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}
}

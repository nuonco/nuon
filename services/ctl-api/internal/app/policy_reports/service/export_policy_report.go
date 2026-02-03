package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type PolicyReportExportFormat string

const (
	PolicyReportExportFormatOPA   PolicyReportExportFormat = "opa"
	PolicyReportExportFormatSARIF PolicyReportExportFormat = "sarif"
	PolicyReportExportFormatPDF   PolicyReportExportFormat = "pdf"
)

const (
	SARIFVersion   = "2.1.0"
	SARIFSchemaURI = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"
)

type PolicyReportTemplateData struct {
	Title         string
	GeneratedAt   string
	ReportID      string
	OrgID         string
	DenyCount     int
	WarnCount     int
	PassCount     int
	TotalCount    int
	Status        string
	Violations    []PolicyViolationDisplay
	HasViolations bool
}

type PolicyViolationDisplay struct {
	PolicyID   string
	Message    string
	Severity   string
	InputIndex int
}

// SARIF types for export
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
	Level   string       `json:"level"`
	Message SARIFMessage `json:"message"`
}

type SARIFMessage struct {
	Text string `json:"text"`
}

// OPA report format for export
type OPAReport struct {
	EvaluatedAt time.Time             `json:"evaluated_at"`
	Violations  []app.PolicyViolation `json:"violations"`
	PolicyIDs   []string              `json:"policy_ids"`
	Policies    []app.PolicyResult    `json:"policies"`
	Inputs      []app.PolicyInputRef  `json:"inputs,omitempty"`
	DenyCount   int                   `json:"deny_count"`
	WarnCount   int                   `json:"warn_count"`
	PassCount   int                   `json:"pass_count"`
}

// @ID						ExportPolicyReport
// @Summary				export policy report
// @Description.markdown	export_policy_report.md
// @Param					report_id	path	string	true	"policy report ID"
// @Param					format		query	string	false	"export format: opa, sarif, pdf (default: opa)"
// @Tags					policy-reports
// @Accept					json
// @Produce				json
// @Produce				application/pdf
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	interface{}
// @Router					/v1/policy-reports/{report_id}/export [get]
func (s *service) ExportPolicyReport(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	reportID := ctx.Param("report_id")
	format := PolicyReportExportFormat(ctx.DefaultQuery("format", string(PolicyReportExportFormatOPA)))

	if !isValidPolicyReportFormat(format) {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid format: %s", format),
			Description: "Valid formats are: opa, sarif, pdf",
		})
		return
	}

	report, err := s.getPolicyReport(ctx, org.ID, reportID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get policy report"))
		return
	}

	switch format {
	case PolicyReportExportFormatPDF:
		s.servePDFReport(ctx, report)
	case PolicyReportExportFormatSARIF:
		s.serveSARIFReport(ctx, report)
	default:
		s.serveOPAReport(ctx, report)
	}
}

func isValidPolicyReportFormat(format PolicyReportExportFormat) bool {
	switch format {
	case PolicyReportExportFormatOPA, PolicyReportExportFormatSARIF, PolicyReportExportFormatPDF:
		return true
	default:
		return false
	}
}

func toOPAFormat(report *app.PolicyReport) OPAReport {
	return OPAReport{
		EvaluatedAt: report.EvaluatedAt,
		Violations:  report.Violations,
		PolicyIDs:   report.PolicyIDs,
		Policies:    report.Policies,
		Inputs:      report.Inputs,
		DenyCount:   report.DenyCount,
		WarnCount:   report.WarnCount,
		PassCount:   report.PassCount,
	}
}

func toSARIFFormat(report *app.PolicyReport) SARIFReport {
	policyIDs := make(map[string]bool)
	for _, v := range report.Violations {
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

	results := make([]SARIFResult, len(report.Violations))
	for i, v := range report.Violations {
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

func (s *service) serveOPAReport(ctx *gin.Context, report *app.PolicyReport) {
	opaReport := toOPAFormat(report)
	content, err := json.Marshal(opaReport)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to marshal OPA report"))
		return
	}

	filename := fmt.Sprintf("policy-report-%s.opa.json", report.ID)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Data(http.StatusOK, "application/json", content)
}

func (s *service) serveSARIFReport(ctx *gin.Context, report *app.PolicyReport) {
	sarifReport := toSARIFFormat(report)
	content, err := json.Marshal(sarifReport)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to marshal SARIF report"))
		return
	}

	filename := fmt.Sprintf("policy-report-%s.sarif.json", report.ID)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Data(http.StatusOK, "application/json", content)
}

func (s *service) servePDFReport(ctx *gin.Context, report *app.PolicyReport) {
	violations := make([]PolicyViolationDisplay, len(report.Violations))
	for i, v := range report.Violations {
		violations[i] = PolicyViolationDisplay{
			PolicyID:   v.PolicyID,
			Message:    v.Message,
			Severity:   v.Severity,
			InputIndex: v.InputIndex,
		}
	}

	status := "passed"
	if report.DenyCount > 0 {
		status = "failed"
	} else if report.WarnCount > 0 {
		status = "warning"
	}

	data := PolicyReportTemplateData{
		Title:         "Policy Report",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		ReportID:      report.ID,
		OrgID:         report.OrgID,
		DenyCount:     report.DenyCount,
		WarnCount:     report.WarnCount,
		PassCount:     report.PassCount,
		TotalCount:    report.DenyCount + report.WarnCount + report.PassCount,
		Status:        status,
		Violations:    violations,
		HasViolations: len(violations) > 0,
	}

	if err := s.renderPDFReport(ctx, data); err != nil {
		ctx.Error(err)
	}
}

func (s *service) renderPDFReport(ctx *gin.Context, data PolicyReportTemplateData) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(data.Title, false)
	pdf.SetAuthor("Nuon", false)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(0, 10, data.Title)
	if pdf.Error() != nil {
		return errors.Wrap(pdf.Error(), "unable to render pdf header")
	}

	pdf.Ln(8)
	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Report ID: %s", data.ReportID))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Organization: %s", data.OrgID))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", data.GeneratedAt))
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 7, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Denies: %d", data.DenyCount))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Warnings: %d", data.WarnCount))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Passes: %d", data.PassCount))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Total: %d", data.TotalCount))
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 7, "Violations")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "", 10)

	if !data.HasViolations {
		pdf.Cell(0, 6, "No policy violations detected.")
	} else {
		for _, v := range data.Violations {
			line := fmt.Sprintf("[%s] %s", v.Severity, v.Message)
			pdf.MultiCell(0, 5, line, "", "L", false)
			pdf.SetFont("Helvetica", "", 9)
			pdf.Cell(0, 5, fmt.Sprintf("Policy: %s | Input: %d", v.PolicyID, v.InputIndex))
			pdf.Ln(5)
			pdf.SetFont("Helvetica", "", 10)
			pdf.Ln(2)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return errors.Wrap(err, "unable to generate pdf")
	}

	filename := fmt.Sprintf("policy-report-%s.pdf", data.ReportID)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Data(http.StatusOK, "application/pdf", buf.Bytes())
	return nil
}

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
	RuleID     string
	Message    string
	Severity   string
	InputIndex int
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

	if format == PolicyReportExportFormatPDF {
		s.servePDFReport(ctx, report)
		return
	}

	if err := s.serveJSONReport(ctx, report, format); err != nil {
		ctx.Error(err)
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

func (s *service) serveJSONReport(ctx *gin.Context, report *app.PolicyReport, format PolicyReportExportFormat) error {
	var targetFormat app.PolicyReportFormat
	if format == PolicyReportExportFormatOPA {
		targetFormat = app.PolicyReportFormatOPA
	} else {
		targetFormat = app.PolicyReportFormatSARIF
	}

	if report.Format != targetFormat {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("report %s is not in %s format", report.ID, format),
			Description: fmt.Sprintf("Report %s is not available in %s format.", report.ID, format),
		})
		return nil
	}

	filename := fmt.Sprintf("policy-report-%s.%s.json", report.ID, format)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Data(http.StatusOK, "application/json", report.Content)
	return nil
}

func (s *service) servePDFReport(ctx *gin.Context, report *app.PolicyReport) {
	if report.Format != app.PolicyReportFormatOPA {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("report %s is not in opa format", report.ID),
			Description: "PDF export requires the OPA report format.",
		})
		return
	}

	var opaData struct {
		EvaluatedAt time.Time `json:"evaluated_at"`
		Violations  []struct {
			PolicyID   string `json:"policy_id"`
			RuleID     string `json:"rule_id"`
			InputIndex int    `json:"input_index"`
			Message    string `json:"message"`
			Severity   string `json:"severity"`
		} `json:"violations"`
		DenyCount int `json:"deny_count"`
		WarnCount int `json:"warn_count"`
		PassCount int `json:"pass_count"`
	}

	if err := json.Unmarshal(report.Content, &opaData); err != nil {
		ctx.Error(errors.Wrap(err, "unable to parse OPA report content"))
		return
	}

	violations := make([]PolicyViolationDisplay, len(opaData.Violations))
	for i, v := range opaData.Violations {
		violations[i] = PolicyViolationDisplay{
			PolicyID:   v.PolicyID,
			RuleID:     v.RuleID,
			Message:    v.Message,
			Severity:   v.Severity,
			InputIndex: v.InputIndex,
		}
	}

	status := "passed"
	if opaData.DenyCount > 0 {
		status = "failed"
	} else if opaData.WarnCount > 0 {
		status = "warning"
	}

	data := PolicyReportTemplateData{
		Title:         "Policy Report",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		ReportID:      report.ID,
		OrgID:         report.OrgID,
		DenyCount:     opaData.DenyCount,
		WarnCount:     opaData.WarnCount,
		PassCount:     opaData.PassCount,
		TotalCount:    opaData.DenyCount + opaData.WarnCount + opaData.PassCount,
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
			if v.RuleID != "" {
				pdf.SetFont("Helvetica", "", 9)
				pdf.Cell(0, 5, fmt.Sprintf("Policy: %s | Rule: %s | Input: %d", v.PolicyID, v.RuleID, v.InputIndex))
				pdf.Ln(5)
				pdf.SetFont("Helvetica", "", 10)
			}
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

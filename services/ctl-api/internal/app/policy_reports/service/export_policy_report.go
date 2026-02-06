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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type PolicyReportExportFormat string

const (
	PolicyReportExportFormatJSON  PolicyReportExportFormat = "json"
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
	OrgName       string // Human-readable name when available
	AppID         string
	AppName       string // Human-readable name when available
	DenyCount     int
	WarnCount     int
	PassCount     int
	TotalCount    int
	Status        string
	Violations    []helpers.PolicyViolationDisplay
	HasViolations bool
	Policies      []helpers.PolicyResultDisplay
	Inputs        []helpers.PolicyInputDisplay
}

// Type aliases for backward compatibility and local convenience
type PolicyResultDisplay = helpers.PolicyResultDisplay
type PolicyInputDisplay = helpers.PolicyInputDisplay
type PolicyViolationDisplay = helpers.PolicyViolationDisplay

// SARIF types for export
type SARIFPropertyBag map[string]any

type SARIFReport struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool        SARIFTool         `json:"tool"`
	Invocations []SARIFInvocation `json:"invocations,omitempty"`
	Artifacts   []SARIFArtifact   `json:"artifacts,omitempty"`
	Results     []SARIFResult     `json:"results"`
	Properties  SARIFPropertyBag  `json:"properties,omitempty"`
}

type SARIFInvocation struct {
	StartTimeUTC        *time.Time       `json:"startTimeUtc,omitempty"`
	EndTimeUTC          *time.Time       `json:"endTimeUtc,omitempty"`
	ExecutionSuccessful bool             `json:"executionSuccessful"`
	Properties          SARIFPropertyBag `json:"properties,omitempty"`
}

type SARIFArtifact struct {
	Location   SARIFArtifactLocation `json:"location"`
	Roles      []string              `json:"roles,omitempty"`
	Properties SARIFPropertyBag      `json:"properties,omitempty"`
}

type SARIFArtifactLocation struct {
	URI   string `json:"uri"`
	Index *int   `json:"index,omitempty"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
}

type SARIFLocation struct {
	PhysicalLocation *SARIFPhysicalLocation `json:"physicalLocation,omitempty"`
	Properties       SARIFPropertyBag       `json:"properties,omitempty"`
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
	ID                   string                       `json:"id"`
	ShortDescription     SARIFMessage                 `json:"shortDescription,omitempty"`
	DefaultConfiguration *SARIFReportingConfiguration `json:"defaultConfiguration,omitempty"`
	Properties           SARIFPropertyBag             `json:"properties,omitempty"`
}

type SARIFReportingConfiguration struct {
	Level   string `json:"level,omitempty"`
	Enabled bool   `json:"enabled"`
}

type SARIFResult struct {
	RuleID     string           `json:"ruleId"`
	Level      string           `json:"level"`
	Message    SARIFMessage     `json:"message"`
	Locations  []SARIFLocation  `json:"locations,omitempty"`
	Properties SARIFPropertyBag `json:"properties,omitempty"`
}

type SARIFMessage struct {
	Text string `json:"text"`
}

// PolicyReportJSON is the JSON export format for policy reports
type PolicyReportJSON struct {
	ReportID    string                `json:"report_id"`
	OrgID       string                `json:"org_id"`
	OrgName     string                `json:"org_name,omitempty"`
	AppID       string                `json:"app_id"`
	AppName     string                `json:"app_name,omitempty"`
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
// @Param					format		query	string	false	"export format: json, sarif, pdf (default: json)"
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
	format := PolicyReportExportFormat(ctx.DefaultQuery("format", string(PolicyReportExportFormatJSON)))

	if !isValidPolicyReportFormat(format) {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid format: %s", format),
			Description: "Valid formats are: json, sarif, pdf",
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
		s.serveJSONReport(ctx, report)
	}
}

func isValidPolicyReportFormat(format PolicyReportExportFormat) bool {
	switch format {
	case PolicyReportExportFormatJSON, PolicyReportExportFormatSARIF, PolicyReportExportFormatPDF:
		return true
	default:
		return false
	}
}

func toJSONFormat(report *app.PolicyReport) PolicyReportJSON {
	return PolicyReportJSON{
		ReportID:    report.ID,
		OrgID:       report.OrgID,
		OrgName:     report.OrgName,
		AppID:       report.AppID,
		AppName:     report.AppName,
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
	// Build artifacts from inputs
	artifacts := make([]SARIFArtifact, len(report.Inputs))
	for i, input := range report.Inputs {
		artifacts[i] = SARIFArtifact{
			Location: SARIFArtifactLocation{
				URI: fmt.Sprintf("nuon://policy-input/%s/%s", input.Type, input.ID),
			},
			Roles: []string{"analysisTarget"},
			Properties: SARIFPropertyBag{
				"input_id":   input.ID,
				"input_type": input.Type,
				"input_name": input.Name,
			},
		}
	}

	// Build policy lookup for per-policy counts
	policyLookup := make(map[string]app.PolicyResult)
	for _, p := range report.Policies {
		policyLookup[p.PolicyID] = p
	}

	// Build rules for ALL policies (not just those with violations)
	rules := make([]SARIFRule, len(report.PolicyIDs))
	for i, policyID := range report.PolicyIDs {
		rule := SARIFRule{
			ID: policyID,
			ShortDescription: SARIFMessage{
				Text: "Policy: " + policyID,
			},
			DefaultConfiguration: &SARIFReportingConfiguration{
				Enabled: true,
			},
		}

		// Add per-policy counts from report.Policies if available
		if policyResult, ok := policyLookup[policyID]; ok {
			rule.ShortDescription.Text = "Policy: " + policyResult.PolicyName
			if policyResult.PolicyName == "" {
				rule.ShortDescription.Text = "Policy: " + policyID
			}
			rule.Properties = SARIFPropertyBag{
				"policy_id":   policyID,
				"policy_name": policyResult.PolicyName,
				"status":      policyResult.Status,
				"deny_count":  policyResult.DenyCount,
				"warn_count":  policyResult.WarnCount,
				"pass_count":  policyResult.PassCount,
				"input_count": policyResult.InputCount,
			}

			// Set default level based on status
			switch policyResult.Status {
			case "deny":
				rule.DefaultConfiguration.Level = "error"
			case "warn":
				rule.DefaultConfiguration.Level = "warning"
			default:
				rule.DefaultConfiguration.Level = "note"
			}
		}

		rules[i] = rule
	}

	// Build results with locations referencing artifacts
	results := make([]SARIFResult, len(report.Violations))
	for i, v := range report.Violations {
		level := "warning"
		if v.Severity == "deny" {
			level = "error"
		}

		result := SARIFResult{
			RuleID: v.PolicyID,
			Level:  level,
			Message: SARIFMessage{
				Text: v.Message,
			},
		}

		// Add location referencing the artifact by index
		if v.InputIndex >= 0 && v.InputIndex < len(artifacts) {
			idx := v.InputIndex
			result.Locations = []SARIFLocation{
				{
					PhysicalLocation: &SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{
							Index: &idx,
						},
					},
				},
			}
		}

		// Preserve InputIdentity and other violation data in properties
		if v.InputIdentity != "" || v.InputIndex >= 0 {
			result.Properties = SARIFPropertyBag{
				"input_index":    v.InputIndex,
				"input_identity": v.InputIdentity,
			}
		}

		results[i] = result
	}

	// Build invocation with timestamp
	evaluatedAt := report.EvaluatedAt
	invocations := []SARIFInvocation{
		{
			EndTimeUTC:          &evaluatedAt,
			ExecutionSuccessful: true,
		},
	}

	// Run-level properties with summary counts
	runProperties := SARIFPropertyBag{
		"deny_count":       report.DenyCount,
		"warn_count":       report.WarnCount,
		"pass_count":       report.PassCount,
		"total_violations": len(report.Violations),
		"total_inputs":     len(report.Inputs),
		"total_policies":   len(report.PolicyIDs),
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
				Invocations: invocations,
				Artifacts:   artifacts,
				Results:     results,
				Properties:  runProperties,
			},
		},
	}
}

func (s *service) serveJSONReport(ctx *gin.Context, report *app.PolicyReport) {
	jsonReport := toJSONFormat(report)
	content, err := json.Marshal(jsonReport)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to marshal JSON report"))
		return
	}

	filename := fmt.Sprintf("policy-report-%s.json", report.ID)
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
	violations := helpers.ToViolationDisplays(report.Violations)
	policies := helpers.ToResultDisplays(report.Policies)
	inputs := helpers.ToInputDisplays(report.Inputs)

	computedDenyCount := 0
	computedWarnCount := 0
	computedPassCount := 0
	for _, p := range policies {
		computedDenyCount += p.DenyCount
		computedWarnCount += p.WarnCount
		computedPassCount += p.PassCount
	}

	status := "passed"
	if report.DenyCount > 0 {
		status = "failed"
	} else if report.WarnCount > 0 {
		status = "warning"
	}

	// Use computed counts from policies for accurate summary
	denyCount := computedDenyCount
	warnCount := computedWarnCount
	passCount := computedPassCount
	// Fall back to report-level counts if no policies data
	if len(report.Policies) == 0 {
		denyCount = report.DenyCount
		warnCount = report.WarnCount
		passCount = report.PassCount
	}

	data := PolicyReportTemplateData{
		Title:         "Policy Report",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		ReportID:      report.ID,
		OrgID:         report.OrgID,
		OrgName:       report.OrgName,
		AppID:         report.AppID,
		AppName:       report.AppName,
		DenyCount:     denyCount,
		WarnCount:     warnCount,
		PassCount:     passCount,
		TotalCount:    denyCount + warnCount + passCount,
		Status:        status,
		Violations:    violations,
		HasViolations: len(violations) > 0,
		Policies:      policies,
		Inputs:        inputs,
	}

	if err := s.renderPDFReport(ctx, data); err != nil {
		ctx.Error(err)
	}
}

// Color constants for severity
var (
	colorDeny = [3]int{200, 50, 50}   // Red
	colorWarn = [3]int{200, 150, 0}   // Yellow/Orange
	colorPass = [3]int{50, 150, 50}   // Green
	colorText = [3]int{40, 40, 40}    // Dark gray
	colorMute = [3]int{120, 120, 120} // Muted gray
)

func (s *service) renderPDFReport(ctx *gin.Context, data PolicyReportTemplateData) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(data.Title, false)
	pdf.SetAuthor("Nuon", false)
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header with title and status badge
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
	pdf.Cell(140, 10, data.Title)

	// Status badge
	statusColor := colorPass
	if data.Status == "failed" {
		statusColor = colorDeny
	} else if data.Status == "warning" {
		statusColor = colorWarn
	}
	pdf.SetTextColor(statusColor[0], statusColor[1], statusColor[2])
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 10, fmt.Sprintf("[%s]", data.Status))
	pdf.Ln(12)

	if pdf.Error() != nil {
		return errors.Wrap(pdf.Error(), "unable to render pdf header")
	}

	// Report metadata
	pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
	pdf.SetFont("Helvetica", "", 9)
	pdf.Cell(0, 5, fmt.Sprintf("Report ID: %s    Generated: %s UTC", data.ReportID, data.GeneratedAt))
	pdf.Ln(5)

	orgDisplay := data.OrgID
	if data.OrgName != "" {
		orgDisplay = fmt.Sprintf("%s (%s)", data.OrgName, data.OrgID)
	}
	pdf.Cell(0, 5, fmt.Sprintf("Organization: %s", orgDisplay))
	pdf.Ln(5)

	appDisplay := data.AppID
	if data.AppName != "" {
		appDisplay = fmt.Sprintf("%s (%s)", data.AppName, data.AppID)
	}
	pdf.Cell(0, 5, fmt.Sprintf("App: %s", appDisplay))
	pdf.Ln(10)

	// Summary section - horizontal layout
	pdf.SetFont("Helvetica", "B", 11)
	pdf.Cell(0, 6, "Summary")
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(colorDeny[0], colorDeny[1], colorDeny[2])
	pdf.Cell(40, 5, fmt.Sprintf("Denies: %d", data.DenyCount))
	pdf.SetTextColor(colorWarn[0], colorWarn[1], colorWarn[2])
	pdf.Cell(40, 5, fmt.Sprintf("Warnings: %d", data.WarnCount))
	pdf.SetTextColor(colorPass[0], colorPass[1], colorPass[2])
	pdf.Cell(40, 5, fmt.Sprintf("Passes: %d", data.PassCount))
	pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
	pdf.Cell(0, 5, fmt.Sprintf("Total: %d", data.TotalCount))
	pdf.Ln(10)

	// Policies Evaluated section
	if len(data.Policies) > 0 {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.Cell(0, 6, "Policies Evaluated")
		pdf.Ln(6)

		// Table header
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(80, 6, "Policy", "1", 0, "L", true, 0, "")
		pdf.CellFormat(25, 6, "Status", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 6, "Denies", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 6, "Warnings", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 6, "Passes", "1", 0, "C", true, 0, "")
		pdf.Ln(6)

		pdf.SetFont("Helvetica", "", 9)
		for _, p := range data.Policies {
			policyDisplay := p.PolicyID
			if p.PolicyName != "" {
				policyDisplay = p.PolicyName
			}

			pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
			pdf.CellFormat(80, 5, truncateString(policyDisplay, 40), "1", 0, "L", false, 0, "")

			// Status with color
			statusColor := getStatusColor(p.Status)
			pdf.SetTextColor(statusColor[0], statusColor[1], statusColor[2])
			pdf.CellFormat(25, 5, p.Status, "1", 0, "C", false, 0, "")

			// Counts
			pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
			pdf.CellFormat(25, 5, fmt.Sprintf("%d", p.DenyCount), "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 5, fmt.Sprintf("%d", p.WarnCount), "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 5, fmt.Sprintf("%d", p.PassCount), "1", 0, "C", false, 0, "")
			pdf.Ln(5)
		}
		pdf.Ln(5)
	}

	// Inputs Evaluated section
	if len(data.Inputs) > 0 {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.Cell(0, 6, "Inputs Evaluated")
		pdf.Ln(6)

		pdf.SetFont("Helvetica", "", 9)
		for _, inp := range data.Inputs {
			inputDisplay := inp.ID
			if inp.Name != "" {
				inputDisplay = fmt.Sprintf("%s (%s)", inp.Name, inp.ID)
			}
			pdf.SetTextColor(colorMute[0], colorMute[1], colorMute[2])
			pdf.Cell(25, 5, inp.Type)
			pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
			pdf.Cell(0, 5, truncateString(inputDisplay, 60))
			pdf.Ln(5)
		}
		pdf.Ln(5)
	}

	// Violations section
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
	pdf.Cell(0, 6, "Violations")
	pdf.Ln(6)

	if !data.HasViolations {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.SetTextColor(colorPass[0], colorPass[1], colorPass[2])
		pdf.Cell(0, 5, "No policy violations detected.")
		pdf.Ln(8)
	} else {
		pdf.SetFont("Helvetica", "", 9)
		for _, v := range data.Violations {
			// Severity indicator with color
			severityColor := getStatusColor(v.Severity)
			pdf.SetTextColor(severityColor[0], severityColor[1], severityColor[2])
			pdf.SetFont("Helvetica", "B", 9)
			pdf.Cell(15, 5, fmt.Sprintf("[%s]", v.Severity))

			// Message
			pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
			pdf.SetFont("Helvetica", "", 9)
			pdf.MultiCell(0, 5, v.Message, "", "L", false)

			// Policy reference with input identity
			pdf.SetTextColor(colorMute[0], colorMute[1], colorMute[2])
			pdf.SetFont("Helvetica", "", 8)
			inputRef := v.InputIdentity
			if inputRef == "" {
				inputRef = fmt.Sprintf("Input Index: %d", v.InputIndex)
			}
			pdf.Cell(0, 4, fmt.Sprintf("Policy: %s | %s", v.PolicyID, inputRef))
			pdf.Ln(6)
		}
	}

	// CLI Reference section
	pdf.Ln(5)
	s.renderCLIReferenceSection(pdf, data)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return errors.Wrap(err, "unable to generate pdf")
	}

	filename := fmt.Sprintf("policy-report-%s.pdf", data.ReportID)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Data(http.StatusOK, "application/pdf", buf.Bytes())
	return nil
}

func (s *service) renderCLIReferenceSection(pdf *fpdf.Fpdf, data PolicyReportTemplateData) {
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
	pdf.Cell(0, 6, "CLI Reference")
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(colorMute[0], colorMute[1], colorMute[2])
	pdf.Cell(0, 4, "Use these commands to explore details:")
	pdf.Ln(5)

	commands := []string{
		fmt.Sprintf("nuon policies reports get -r %s", data.ReportID),
		fmt.Sprintf("nuon policies get -a %s", data.AppID),
	}

	// Add component/build commands if we have inputs
	for _, inp := range data.Inputs {
		switch inp.Type {
		case "component":
			commands = append(commands, fmt.Sprintf("nuon components get -a %s -c %s", data.AppID, inp.ID))
		case "component_build":
			commands = append(commands, fmt.Sprintf("nuon builds get -a %s -b %s", data.AppID, inp.ID))
			commands = append(commands, fmt.Sprintf("nuon builds logs -a %s -b %s", data.AppID, inp.ID))
		}
	}

	pdf.SetFont("Courier", "", 8)
	pdf.SetTextColor(colorText[0], colorText[1], colorText[2])
	for _, cmd := range commands {
		pdf.Cell(5, 4, "")
		pdf.Cell(0, 4, cmd)
		pdf.Ln(4)
	}
}

func getStatusColor(status string) [3]int {
	switch status {
	case "deny":
		return colorDeny
	case "warn":
		return colorWarn
	case "pass":
		return colorPass
	default:
		return colorText
	}
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

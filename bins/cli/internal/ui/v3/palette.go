package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
)

// NOTE: contributed by claude
func main() {
	fmt.Println()
	fmt.Println(styles.TextBold.Render("🎨 NUON CLI LIPGLOSS STYLE PALETTE"))
	fmt.Println()

	// Color Palette Section
	fmt.Println(styles.TextBold.Render("📍 COLOR PALETTE"))
	fmt.Println()

	colorTable := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(styles.BorderColor)).
		Headers("Color", "Preview", "Usage").
		Width(80)

	colorTable.Row("Primary", styles.TextBold.Background(styles.PrimaryColor).Render("  Primary  "), "Main brand color")
	colorTable.Row("Secondary", styles.TextBold.Background(styles.SecondaryColor).Render("  Secondary  "), "Accent color")
	colorTable.Row("Accent", styles.TextBold.Background(styles.AccentColor).Render("  Accent  "), "Highlight color")
	colorTable.Row("Dim", styles.TextBold.Background(styles.Dim).Render("  Dim  "), "Muted elements")
	colorTable.Row("Ghost", styles.TextBold.Background(styles.Ghost).Render("  Ghost  "), "Adaptive subtle")

	fmt.Println(colorTable.Render())
	fmt.Println()

	// Text Styles Section
	fmt.Println(styles.TextBold.Render("✨ TEXT STYLES"))
	fmt.Println()

	textStyles := []struct {
		name    string
		style   lipgloss.Style
		example string
	}{
		{"Link", styles.Link, "https://example.com"},
		{"Text", styles.Text, "Default text style"},
		{"Text Ghost", styles.TextGhost, "Subtle italic text"},
		{"Text Bold", styles.TextBold, "Bold text for emphasis"},
		{"Text Dim", styles.TextDim, "Dimmed text for less important content"},
		{"Text Light", styles.TextLight, "Light colored text"},
		{"Text Success", styles.TextSuccess, "Success messages"},
		{"Text Error", styles.TextError, "Error messages"},
	}

	textTable := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(styles.BorderColor)).
		Headers("Style Name", "Example").
		Width(80)

	for _, ts := range textStyles {
		textTable.Row(ts.name, ts.style.Render(ts.example))
	}

	fmt.Println(textTable.Render())
	fmt.Println()

	// Status Styles Section
	fmt.Println(styles.TextBold.Render("🚦 STATUS STYLES"))
	fmt.Println()

	statusTable := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(styles.BorderColor)).
		Headers("Status", "Style", "Description").
		Width(80)

	statuses := []struct {
		status models.AppStatus
		desc   string
	}{
		{models.AppStatusPending, "Awaiting processing"},
		{models.AppStatusNotDashAttempted, "Not yet started"},
		{models.AppStatusApproved, "Successfully approved"},
		{models.AppStatusApprovalDashDenied, "Approval was denied"},
		{models.AppStatusCancelled, "Operation cancelled"},
		{models.AppStatusError, "Error occurred"},
		{models.AppStatusAutoDashSkipped, "Automatically skipped"},
		{models.AppStatusApprovalDashAwaiting, "Waiting for approval"},
		{models.AppStatusSuccess, "Successfully completed"},
	}

	for _, s := range statuses {
		style := styles.GetStatusStyle(s.status)
		statusTable.Row(string(s.status), style.Render(string(s.status)), s.desc)
	}

	fmt.Println(statusTable.Render())
	fmt.Println()

	// Banner Styles Section
	fmt.Println(styles.TextBold.Render("🎯 BANNER STYLES"))
	fmt.Println()

	fmt.Println("Approval Confirmation:")
	fmt.Println(styles.ApprovalConfirmation.Render("✅ Your deployment has been approved and is ready to proceed!"))
	fmt.Println()

	fmt.Println("Success Banner:")
	fmt.Println(styles.SuccessBanner.Render("🎉 Deployment completed successfully!"))
	fmt.Println()

	// Log Message Style
	fmt.Println(styles.TextBold.Render("📋 LOG MESSAGE STYLE"))
	fmt.Println()
	fmt.Println(styles.LogMessageStyle.Render("2024-01-01 12:00:00 INFO: Application started successfully"))
	fmt.Println(styles.LogMessageStyle.Render("2024-01-01 12:00:01 DEBUG: Processing user request"))
	fmt.Println(styles.LogMessageStyle.Render("2024-01-01 12:00:02 WARN: High memory usage detected"))
	fmt.Println()

	// Help Style
	fmt.Println(styles.TextBold.Render("❓ HELP STYLE"))
	fmt.Println()
	helpContent := `This is an example of help content that would be displayed to users.
It includes multiple lines of text with proper padding and formatting.
Use this style for help messages and documentation.`
	fmt.Println(styles.HelpStyle.Render(helpContent))
	fmt.Println()

	// Common Elements Section
	fmt.Println(styles.TextBold.Render("🧩 COMMON ELEMENTS"))
	fmt.Println()

	// Humanized Duration Examples
	fmt.Println(styles.TextBold.Render("Duration Formatting:"))
	durations := []int64{0, 5000000000, 90000000000, 3661000000000} // 0s, 5s, 1m30s, 1h1m1s
	for _, d := range durations {
		fmt.Printf("  %d ns → %s\n", d, common.HumanizeNSDuration(d))
	}
	fmt.Println()

	// Full Page Dialog Examples
	fmt.Println(styles.TextBold.Render("Full Page Dialogs:"))
	fmt.Println()

	// Info dialog
	infoDialog := common.FullPageDialog(common.FullPageDialogRequest{
		Width:   60,
		Height:  10,
		Content: "ℹ️  This is an info dialog\nShowing important information",
		Level:   "info",
	})
	fmt.Println("Info Dialog:")
	fmt.Println(infoDialog)

	// Warning dialog
	warningDialog := common.FullPageDialog(common.FullPageDialogRequest{
		Width:   60,
		Height:  8,
		Content: "⚠️  Warning: This action cannot be undone",
		Level:   "warning",
	})
	fmt.Println("Warning Dialog:")
	fmt.Println(warningDialog)

	// Error dialog
	errorDialog := common.FullPageDialog(common.FullPageDialogRequest{
		Width:   60,
		Height:  8,
		Content: "❌ Error: Operation failed",
		Level:   "error",
	})
	fmt.Println("Error Dialog:")
	fmt.Println(errorDialog)

	// Complex Table Example
	fmt.Println(styles.TextBold.Render("📊 COMPLEX TABLE EXAMPLE"))
	fmt.Println()

	complexTable := table.New().
		Border(lipgloss.ThickBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(styles.PrimaryColor)).
		Headers("App Name", "Status", "Last Deploy", "Health").
		Width(100)

	apps := []struct {
		name       string
		status     models.AppStatus
		lastDeploy string
		health     string
	}{
		{"web-frontend", models.AppStatusSuccess, "2h ago", "Healthy"},
		{"api-backend", models.AppStatusPending, "5m ago", "Deploying"},
		{"worker-queue", models.AppStatusError, "1d ago", "Failed"},
		{"database", models.AppStatusApproved, "3h ago", "Healthy"},
	}

	for _, app := range apps {
		statusStyle := styles.GetStatusStyle(app.status)
		var healthStyle lipgloss.Style
		switch app.health {
		case "Healthy":
			healthStyle = styles.TextSuccess
		case "Deploying":
			healthStyle = styles.Info
		case "Failed":
			healthStyle = styles.Error
		default:
			healthStyle = styles.TextDim
		}

		complexTable.Row(
			styles.TextBold.Render(app.name),
			statusStyle.Render(string(app.status)),
			styles.TextDim.Render(app.lastDeploy),
			healthStyle.Render(app.health),
		)
	}

	fmt.Println(complexTable.Render())
	fmt.Println()

	// Layout Examples
	fmt.Println(styles.TextBold.Render("📐 LAYOUT EXAMPLES"))
	fmt.Println()

	// Side-by-side content
	leftContent := lipgloss.NewStyle().
		Width(30).
		Align(lipgloss.Left).
		Foreground(styles.SecondaryColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1).
		Render("Left Panel\n\nThis content is\naligned to the left\nside of the layout.")

	rightContent := lipgloss.NewStyle().
		Width(30).
		Align(lipgloss.Right).
		Foreground(styles.AccentColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1).
		Render("Right Panel\n\nThis content is\naligned to the right\nside of the layout.")

	sideBySide := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, "  ", rightContent)
	fmt.Println(sideBySide)
	fmt.Println()

	// Centered content with border
	centeredStyle := lipgloss.NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Foreground(styles.PrimaryColor).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(styles.PrimaryColor).
		Padding(1, 2)

	centered := centeredStyle.Render(strings.Join([]string{
		"🎯 CENTERED CONTENT",
		"",
		"This text is perfectly centered",
		"within a bordered container",
		"using lipgloss styling",
	}, "\n"))

	fmt.Println(centered)
	fmt.Println()

	fmt.Println(styles.TextSuccess.Render("✅ Style palette demonstration complete!"))
	fmt.Println(styles.TextDim.Render("This sampler showcases all available styles and components."))
	fmt.Println()
}

package bubbles

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	debugEnvVar                = "NUON_DEBUG"
	defaultServerErrorMessage  = "Oops, we have experienced a server error. Please try again in a few minutes."
	defaultUnknownErrorMessage = "Oops, we have experienced an unexpected error. Please let us know about this."
)

// CLI User error for custom CLI errors
type CLIUserError struct {
	Msg string
}

func (u *CLIUserError) Error() string {
	return u.Msg
}

// Message functions using Lipgloss styling

// PrintError prints error messages with appropriate styling and error handling logic
func PrintError(err error) error {
	if os.Getenv(debugEnvVar) != "" {
		debugStyle := lipgloss.NewStyle().
			Foreground(SubtleColor).
			Italic(true)
		fmt.Println(debugStyle.Render("DEBUG: " + err.Error()))
	}

	// Construct a stack trace if this error doesn't already have one
	if !errs.HasNuonStackTrace(err) {
		err = errors.WithStackDepth(err, 1)
	}

	cliUserErr := &CLIUserError{}
	if errors.As(err, &cliUserErr) {
		fmt.Println(ErrorStyle.Render("✗ " + err.Error()))
		return err
	}

	apiUserErr, ok := nuon.ToUserError(err)
	if ok {
		fmt.Println(ErrorStyle.Render("✗ " + apiUserErr.Description))
		return err
	}

	if nuon.IsServerError(err) {
		fmt.Println(ErrorStyle.Render("✗ " + defaultServerErrorMessage))
		return err
	}

	var cfgErr config.ErrConfig
	if errors.As(err, &cfgErr) {
		msg := fmt.Sprintf("%s %s", cfgErr.Description, cfgErr.Error())
		if cfgErr.Warning {
			fmt.Println(WarningStyle.Render("⚠ " + msg))
			return cfgErr
		}

		fmt.Println(ErrorStyle.Render("✗ " + msg))
		return cfgErr
	}

	var syncErr sync.SyncErr
	if errors.As(err, &syncErr) {
		fmt.Println(ErrorStyle.Render("✗ " + syncErr.Error()))
		return syncErr
	}

	var syncAPIErr sync.SyncAPIErr
	if errors.As(err, &syncAPIErr) {
		fmt.Println(ErrorStyle.Render("✗ " + syncAPIErr.Error()))
		return syncAPIErr
	}

	var parseErr parse.ParseErr
	if errors.As(err, &parseErr) {
		fmt.Println(ErrorStyle.Render("✗ " + parseErr.Description))
		if parseErr.Err != nil {
			fmt.Println(ErrorStyle.Render("✗ " + parseErr.Err.Error()))
		}

		return parseErr
	}

	fmt.Println(ErrorStyle.Render("✗ " + err.Error()))
	return err
}

// PrintLn prints informational messages
func PrintLn(msg string) {
	fmt.Println(InfoStyle.Render("ℹ " + msg))
}

// PrintWarning prints warning messages
func PrintWarning(msg string) {
	fmt.Println(WarningStyle.Render("⚠ " + msg))
}

// PrintSuccess prints success messages
func PrintSuccess(msg string) {
	fmt.Println(SuccessStyle.Render("✓ " + msg))
}

// PrintBasicText prints basic text without styling (for backward compatibility)
func PrintBasicText(msg string) {
	fmt.Println(BaseStyle.Render(msg))
}

// PrintWithColor prints colored text with the specified color
func PrintWithColor(msg, color string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	fmt.Println(style.Render(msg))
}

// Evaluation journey specific messages
func PrintEvaluationWelcome(orgName string) {
	welcomeStyle := lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(AccentColor).
		Padding(1, 2).
		Margin(1, 0)

	content := fmt.Sprintf("🚀 Welcome to Nuon Evaluation!\n\nYou're now working with the '%s' organization.\n\nThis evaluation environment is perfect for trying out Nuon's features\nwithout affecting production workloads.", orgName)

	fmt.Println(welcomeStyle.Render(content))
	fmt.Println(EvaluationTipStyle.Render("Use 'nuon help' to explore available commands"))
}

func PrintEvaluationTip(tip string) {
	fmt.Println(EvaluationTipStyle.Render("💡 Tip: " + tip))
}

// JSON output functions
func PrintJSON(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonBytes))
	return nil
}

func PrintJSONError(err error) {
	errorData := map[string]string{
		"error": err.Error(),
	}
	PrintJSON(errorData)
}

// Formatted printing functions
func PrintFormattedMessage(level, prefix, message string, color lipgloss.Color) {
	style := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Padding(0, 1)
	
	fmt.Println(style.Render(prefix + " " + message))
}

func PrintfLn(format string, args ...interface{}) {
	PrintLn(fmt.Sprintf(format, args...))
}

func PrintfSuccess(format string, args ...interface{}) {
	PrintSuccess(fmt.Sprintf(format, args...))
}

func PrintfWarning(format string, args ...interface{}) {
	PrintWarning(fmt.Sprintf(format, args...))
}

func PrintfError(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	PrintError(err)
}

// Progress and status indicators
func PrintStep(step int, total int, message string) {
	stepStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Padding(0, 1)
	
	stepText := fmt.Sprintf("[%d/%d] %s", step, total, message)
	fmt.Println(stepStyle.Render(stepText))
}

func PrintDivider() {
	dividerStyle := lipgloss.NewStyle().
		Foreground(BorderColor).
		Margin(1, 0)
	
	divider := strings.Repeat("─", 60)
	fmt.Println(dividerStyle.Render(divider))
}

// Helper functions for backward compatibility
func ColoredText(text, color string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}

func Green(text string) string {
	return ColoredText(text, "#10b981")
}

func LightMagenta(text string) string {
	return ColoredText(text, "#d946ef")
}

func Red(text string) string {
	return ColoredText(text, "#ef4444")
}

func Blue(text string) string {
	return ColoredText(text, "#3b82f6")
}

func Yellow(text string) string {
	return ColoredText(text, "#f59e0b")
}

// Context-aware messaging
func PrintContextMessage(context, message string) {
	contextStyle := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true)
	
	fmt.Println(contextStyle.Render(fmt.Sprintf("[%s] %s", context, message)))
}

func PrintCurrentStatus(entity, name, id string) {
	statusStyle := lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)
	
	message := fmt.Sprintf("✓ current %s is now %s: %s", entity, Green(name), Green(id))
	fmt.Println(statusStyle.Render(message))
}

func PrintUnsetStatus(entity string) {
	statusStyle := lipgloss.NewStyle().
		Foreground(InfoColor).
		Bold(true)
	
	message := fmt.Sprintf("ℹ current %s is now %s", entity, Green("unset"))
	fmt.Println(statusStyle.Render(message))
}

func PrintNotFoundMessage(entity, id, listCommand string) {
	message := fmt.Sprintf("can't find %s %s, use %s to view all %ss", entity, Green(id), LightMagenta(listCommand), entity)
	fmt.Println(WarningStyle.Render(message))
}

func PrintNoItemsMessage(entity, createCommand string) {
	message := fmt.Sprintf("you don't have any %ss, create one using %s", entity, LightMagenta(createCommand))
	fmt.Println(InfoStyle.Render(message))
}

func PrintNotSetMessage(entity, selectCommand string) {
	message := fmt.Sprintf("current %s is not set, use %s to set one", entity, LightMagenta(selectCommand))
	fmt.Println(WarningStyle.Render(message))
}
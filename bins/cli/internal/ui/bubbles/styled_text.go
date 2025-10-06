package bubbles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// StyledText provides consistent text styling throughout the application
type StyledText struct {
	successStyle   lipgloss.Style
	errorStyle     lipgloss.Style
	warningStyle   lipgloss.Style
	infoStyle      lipgloss.Style
	highlightStyle lipgloss.Style
	subtleStyle    lipgloss.Style
}

// NewStyledText creates a new styled text instance
func NewStyledText() *StyledText {
	return &StyledText{
		successStyle:   lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true),
		errorStyle:     lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true),
		warningStyle:   lipgloss.NewStyle().Foreground(styles.WarningColor).Bold(true),
		infoStyle:      lipgloss.NewStyle().Foreground(styles.PrimaryColor),
		highlightStyle: lipgloss.NewStyle().Foreground(styles.AccentColor).Bold(true),
		subtleStyle:    lipgloss.NewStyle().Foreground(styles.SubtleColor),
	}
}

// Success renders text in success style (green, bold)
func (s *StyledText) Success(text string) string {
	return s.successStyle.Render(text)
}

// Error renders text in error style (red, bold)
func (s *StyledText) Error(text string) string {
	return s.errorStyle.Render(text)
}

// Warning renders text in warning style (yellow, bold)
func (s *StyledText) Warning(text string) string {
	return s.warningStyle.Render(text)
}

// Info renders text in info style (primary color)
func (s *StyledText) Info(text string) string {
	return s.infoStyle.Render(text)
}

// Highlight renders text in highlight style (bright, bold)
func (s *StyledText) Highlight(text string) string {
	return s.highlightStyle.Render(text)
}

// Subtle renders text in subtle style (muted)
func (s *StyledText) Subtle(text string) string {
	return s.subtleStyle.Render(text)
}

// Print outputs plain text without styling
func (s *StyledText) Print(text string) {
	fmt.Print(text)
}

// Println outputs plain text with newline
func (s *StyledText) Println(text string) {
	fmt.Println(text)
}

// PrintSuccess outputs success-styled text with newline
func (s *StyledText) PrintSuccess(text string) {
	fmt.Println(s.Success(text))
}

// PrintError outputs error-styled text with newline
func (s *StyledText) PrintError(text string) {
	fmt.Println(s.Error(text))
}

// PrintWarning outputs warning-styled text with newline
func (s *StyledText) PrintWarning(text string) {
	fmt.Println(s.Warning(text))
}

// PrintInfo outputs info-styled text with newline
func (s *StyledText) PrintInfo(text string) {
	fmt.Println(s.Info(text))
}

// PrintHighlight outputs highlight-styled text with newline
func (s *StyledText) PrintHighlight(text string) {
	fmt.Println(s.Highlight(text))
}

// PrintSubtle outputs subtle-styled text with newline
func (s *StyledText) PrintSubtle(text string) {
	fmt.Println(s.Subtle(text))
}

// Package-level styled text instance for global use
var styledText = NewStyledText()

// Global convenience functions that match pterm patterns

// PrintPlain outputs plain text (replacement for pterm.Println)
func PrintPlain(text string) {
	fmt.Println(text)
}

// PrintSuccess outputs success text (replacement for pterm.LightGreen)
func PrintStyledSuccess(text string) {
	styledText.PrintSuccess(text)
}

// PrintError outputs error text (replacement for pterm.LightRed)
func PrintStyledError(text string) {
	styledText.PrintError(text)
}

// PrintInfo outputs info text (replacement for pterm.LightCyan)
func PrintStyledInfo(text string) {
	styledText.PrintInfo(text)
}

// PrintHighlight outputs highlighted text (replacement for pterm.LightMagenta)
func PrintStyledHighlight(text string) {
	styledText.PrintHighlight(text)
}

// StyleSuccess returns success-styled text (replacement for pterm.LightGreen())
func StyleSuccess(text string) string {
	return styledText.Success(text)
}

// StyleError returns error-styled text (replacement for pterm.LightRed())
func StyleError(text string) string {
	return styledText.Error(text)
}

// StyleInfo returns info-styled text (replacement for pterm.LightCyan())
func StyleInfo(text string) string {
	return styledText.Info(text)
}

// StyleHighlight returns highlighted text (replacement for pterm.LightMagenta())
func StyleHighlight(text string) string {
	return styledText.Highlight(text)
}

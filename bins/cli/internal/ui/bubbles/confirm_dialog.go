package bubbles

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

// ConfirmDialogModel represents an interactive confirmation dialog
type ConfirmDialogModel struct {
	message   string
	note      string // optional highlighted note rendered before the question
	confirmed bool
	cancelled bool
	quitting  bool
	cursor    int // 0 = Yes, 1 = No
}

// NewConfirmDialog creates a new confirmation dialog
func NewConfirmDialog(message string) ConfirmDialogModel {
	return ConfirmDialogModel{
		message: message,
		cursor:  0, // Default to "Yes"
	}
}

// NewConfirmDialogWithNote creates a confirmation dialog with an additional highlighted note
func NewConfirmDialogWithNote(message, note string) ConfirmDialogModel {
	return ConfirmDialogModel{
		message: message,
		note:    note,
		cursor:  0,
	}
}

// Init initializes the confirmation dialog
func (m ConfirmDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the confirmation dialog
func (m ConfirmDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			m.quitting = true
			return m, tea.Quit

		case "left", "right", "tab":
			// Toggle between Yes/No
			m.cursor = 1 - m.cursor
			return m, nil

		case "enter", "space":
			if m.cursor == 0 {
				m.confirmed = true
			} else {
				m.cancelled = true
			}
			m.quitting = true
			return m, tea.Quit

		default:
			if text := msg.Key().Text; text != "" {
				switch strings.ToLower(text) {
				case "y", "yes":
					m.confirmed = true
					m.quitting = true
					return m, tea.Quit
				case "n", "no":
					m.cancelled = true
					m.quitting = true
					return m, tea.Quit
				}
			}
		}
	}

	return m, nil
}

// View renders the confirmation dialog
func (m ConfirmDialogModel) View() tea.View {
	if m.quitting {
		if m.confirmed {
			successStyle := lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
			return tea.NewView(successStyle.Render("✓ Confirmed"))
		}
		return tea.NewView("")
	}

	var b strings.Builder

	// Optional warning note
	if m.note != "" {
		noteStyle := lipgloss.NewStyle().
			Foreground(styles.WarningColor).
			Bold(true).
			Margin(0, 0, 1, 0)
		b.WriteString(noteStyle.Render("⚠ " + m.note))
		b.WriteString("\n")
	}

	// Question
	questionStyle := lipgloss.NewStyle().
		Foreground(styles.PrimaryColor).
		Bold(true).
		Margin(0, 0, 1, 0)
	b.WriteString(questionStyle.Render(m.message))
	b.WriteString("\n")

	// Yes/No buttons
	yesStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.SubtleColor)
	noStyle := yesStyle

	if m.cursor == 0 {
		// Yes is selected
		yesStyle = yesStyle.
			BorderForeground(styles.PrimaryColor).
			Foreground(styles.PrimaryColor).
			Bold(true)
	} else {
		// No is selected
		noStyle = noStyle.
			BorderForeground(styles.PrimaryColor).
			Foreground(styles.PrimaryColor).
			Bold(true)
	}

	yesButton := yesStyle.Render("Yes")
	noButton := noStyle.Render("No")

	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, yesButton, "  ", noButton)
	b.WriteString(buttonRow)
	b.WriteString("\n")

	// Instructions
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.SubtleColor).
		Italic(true).
		Margin(1, 0, 0, 0)
	b.WriteString(helpStyle.Render("Use ←/→ to navigate, Enter to confirm, Esc to cancel, or type y/n"))

	return tea.NewView(BorderStyle.Render(b.String()))
}

// Result returns whether the dialog was confirmed
func (m ConfirmDialogModel) Result() (bool, bool) {
	return m.confirmed, m.cancelled
}

// Show displays the confirmation dialog and returns the result
// This provides a pterm-compatible API for easy migration
func ShowConfirmDialog(message string, interactive bool) (bool, error) {
	return showDialog(NewConfirmDialog(message), interactive)
}

// ShowConfirmDialogWithNote displays a confirmation dialog with a highlighted warning note above the question.
func ShowConfirmDialogWithNote(message, note string, interactive bool) (bool, error) {
	return showDialog(NewConfirmDialogWithNote(message, note), interactive)
}

func showDialog(model ConfirmDialogModel, interactive bool) (bool, error) {
	if !interactive {
		return false, fmt.Errorf("interactive terminal required for confirmation; use --yes flag to auto-approve")
	}

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return false, err
	}

	confirmModel := finalModel.(ConfirmDialogModel)
	confirmed, cancelled := confirmModel.Result()

	if cancelled {
		return false, fmt.Errorf("confirmation cancelled by user")
	}

	return confirmed, nil
}

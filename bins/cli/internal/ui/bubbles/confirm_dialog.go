package bubbles

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// ConfirmDialogModel represents an interactive confirmation dialog
type ConfirmDialogModel struct {
	message   string
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

// Init initializes the confirmation dialog
func (m ConfirmDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the confirmation dialog
func (m ConfirmDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			m.quitting = true
			return m, tea.Quit

		case tea.KeyLeft, tea.KeyRight, tea.KeyTab:
			// Toggle between Yes/No
			m.cursor = 1 - m.cursor
			return m, nil

		case tea.KeyEnter, tea.KeySpace:
			if m.cursor == 0 {
				m.confirmed = true
			} else {
				m.cancelled = true
			}
			m.quitting = true
			return m, tea.Quit

		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				switch strings.ToLower(string(msg.Runes)) {
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
func (m ConfirmDialogModel) View() string {
	if m.quitting {
		if m.confirmed {
			successStyle := lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
			return successStyle.Render("✓ Confirmed")
		}
		return ""
	}

	var b strings.Builder

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
	noStyle := yesStyle.Copy()

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

	return BorderStyle.Render(b.String())
}

// Result returns whether the dialog was confirmed
func (m ConfirmDialogModel) Result() (bool, bool) {
	return m.confirmed, m.cancelled
}

// Show displays the confirmation dialog and returns the result
// This provides a pterm-compatible API for easy migration
func ShowConfirmDialog(message string) (bool, error) {
	model := NewConfirmDialog(message)

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

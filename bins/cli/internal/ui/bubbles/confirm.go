package bubbles

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmModel represents a yes/no confirmation dialog
type ConfirmModel struct {
	prompt   string
	result   bool
	selected bool
	quitting bool
	choice   int // 0 = Yes, 1 = No
}

// NewConfirmModel creates a new confirmation model
func NewConfirmModel(prompt string) ConfirmModel {
	return ConfirmModel{
		prompt: prompt,
		choice: 0, // Default to Yes
	}
}

// Init initializes the confirmation model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the confirmation model
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyLeft, tea.KeyShiftTab:
			if m.choice > 0 {
				m.choice--
			}

		case tea.KeyRight, tea.KeyTab:
			if m.choice < 1 {
				m.choice++
			}

		case tea.KeyEnter, tea.KeySpace:
			m.result = m.choice == 0 // Yes = true, No = false
			m.selected = true
			m.quitting = true
			return m, tea.Quit
		}

		// Handle y/n key shortcuts
		switch strings.ToLower(msg.String()) {
		case "y":
			m.choice = 0
			m.result = true
			m.selected = true
			m.quitting = true
			return m, tea.Quit
		case "n":
			m.choice = 1
			m.result = false
			m.selected = true
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the confirmation dialog
func (m ConfirmModel) View() string {
	if m.quitting {
		if m.selected {
			resultText := "No"
			if m.result {
				resultText = "Yes"
			}
			return SuccessStyle.Render(fmt.Sprintf("✓ %s", resultText))
		}
		return ""
	}

	// Render the prompt
	promptStyle := lipgloss.NewStyle().
		Foreground(TextColor).
		Bold(true).
		Margin(1, 0)

	// Render the options
	yesStyle := BlurredStyle
	noStyle := BlurredStyle

	if m.choice == 0 {
		yesStyle = FocusedStyle
	} else {
		noStyle = FocusedStyle
	}

	yesButton := yesStyle.Render("[ Yes ]")
	noButton := noStyle.Render("[ No ]")

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Left,
		yesButton,
		"  ",
		noButton,
	)

	help := lipgloss.NewStyle().
		Foreground(SubtleColor).
		Italic(true).
		Margin(1, 0).
		Render("Use arrow keys or tab to navigate • Enter to confirm • y/n for shortcuts")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		promptStyle.Render(m.prompt),
		buttons,
		help,
	)

	return BorderStyle.Render(content)
}

// Result returns the confirmation result
func (m ConfirmModel) Result() bool {
	return m.result
}

// Selected returns whether a choice was made
func (m ConfirmModel) Selected() bool {
	return m.selected
}

// High-level confirmation functions

// Confirm shows a yes/no confirmation dialog
func Confirm(prompt string) (bool, error) {
	model := NewConfirmModel(prompt)

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return false, err
	}

	confirmModel := finalModel.(ConfirmModel)
	if !confirmModel.Selected() {
		return false, fmt.Errorf("confirmation cancelled")
	}

	return confirmModel.Result(), nil
}

// PromptWithAutoApprove prompts for confirmation unless auto-approved
func PromptWithAutoApprove(autoApprove bool, msg string, vars ...interface{}) error {
	if autoApprove {
		return nil
	}

	prompt := fmt.Sprintf(msg, vars...)
	result, err := Confirm(prompt)
	if err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("operation cancelled")
	}

	return nil
}

// ConfirmWithDefault shows a confirmation with a default choice
func ConfirmWithDefault(prompt string, defaultYes bool) (bool, error) {
	model := NewConfirmModel(prompt)
	if !defaultYes {
		model.choice = 1 // Default to No
	}

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return defaultYes, err
	}

	confirmModel := finalModel.(ConfirmModel)
	if !confirmModel.Selected() {
		return defaultYes, nil // Return default if cancelled
	}

	return confirmModel.Result(), nil
}

// EvaluationConfirm shows a confirmation with evaluation journey styling
func EvaluationConfirm(prompt string) (bool, error) {
	// Add evaluation-specific styling or messaging
	evaluationPrompt := fmt.Sprintf("🚀 %s\n\n💡 This is part of your Nuon evaluation experience.", prompt)
	return Confirm(evaluationPrompt)
}
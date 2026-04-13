package bubbles

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

// InlineConfirmModel is a compact single-line yes/no prompt
// that renders as: ? Prompt (Y/n)
type InlineConfirmModel struct {
	prompt     string
	defaultYes bool
	result     bool
	selected   bool
	quitting   bool
}

// NewInlineConfirmModel creates a new inline confirmation model
func NewInlineConfirmModel(prompt string, defaultYes bool) InlineConfirmModel {
	return InlineConfirmModel{
		prompt:     prompt,
		defaultYes: defaultYes,
		result:     defaultYes,
	}
}

func (m InlineConfirmModel) Init() tea.Cmd { return nil }

func (m InlineConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "y", "Y":
			m.result = true
			m.selected = true
			m.quitting = true
			return m, tea.Quit
		case "n", "N":
			m.result = false
			m.selected = true
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.result = m.defaultYes
			m.selected = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m InlineConfirmModel) View() tea.View {
	questionMark := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A2C929")).
		Bold(true).
		Render("?")

	promptText := lipgloss.NewStyle().Bold(true).Render(m.prompt)

	if m.quitting && m.selected {
		answer := "No"
		if m.result {
			answer = "Yes"
		}
		answerStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)
		return tea.NewView(fmt.Sprintf("%s %s %s", questionMark, promptText, answerStyle.Render(answer)))
	}

	if m.quitting {
		return tea.NewView("")
	}

	hintStyle := lipgloss.NewStyle().Foreground(styles.SubtleColor)
	boldHint := lipgloss.NewStyle().Foreground(styles.SubtleColor).Bold(true)

	var hint string
	if m.defaultYes {
		hint = fmt.Sprintf("(%s/%s)", boldHint.Render("Y"), hintStyle.Render("n"))
	} else {
		hint = fmt.Sprintf("(%s/%s)", hintStyle.Render("y"), boldHint.Render("N"))
	}

	return tea.NewView(fmt.Sprintf("%s %s %s ", questionMark, promptText, hint))
}

func (m InlineConfirmModel) Result() bool   { return m.result }
func (m InlineConfirmModel) Selected() bool { return m.selected }

// InlineConfirm shows a compact single-line yes/no prompt: ? Prompt (Y/n)
func InlineConfirm(prompt string, defaultYes, interactive bool) (bool, error) {
	if !interactive {
		return false, fmt.Errorf("interactive terminal required for confirmation; use --yes flag to auto-approve")
	}

	model := NewInlineConfirmModel(prompt, defaultYes)
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return defaultYes, err
	}

	m := finalModel.(InlineConfirmModel)
	if !m.Selected() {
		return defaultYes, fmt.Errorf("confirmation cancelled")
	}

	return m.Result(), nil
}

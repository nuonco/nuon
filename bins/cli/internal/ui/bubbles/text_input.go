package bubbles

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

// TextInputModel is a compact inline text input prompt
// that renders as: ? Prompt [? for help]
type TextInputModel struct {
	input    textinput.Model
	prompt   string
	help     string
	required bool
	showHelp bool
	quitting bool
	err      string
}

// NewTextInputModel creates a new inline text input model
func NewTextInputModel(prompt, placeholder, help string, required bool) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 500
	ti.SetWidth(60)
	ti.Prompt = ""
	ti.Focus()

	return TextInputModel{
		input:    ti,
		prompt:   prompt,
		help:     help,
		required: required,
	}
}

func (m TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "?":
			// If input is empty, toggle help instead of typing "?"
			if m.input.Value() == "" && m.help != "" {
				m.showHelp = !m.showHelp
				return m, nil
			}

		case "enter":
			value := strings.TrimSpace(m.input.Value())
			if m.required && value == "" {
				m.err = "This field is required"
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.err = ""
	return m, cmd
}

func (m TextInputModel) View() tea.View {
	questionMark := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A2C929")).
		Bold(true).
		Render("?")
	promptText := lipgloss.NewStyle().Bold(true).Render(m.prompt)

	if m.quitting {
		value := strings.TrimSpace(m.input.Value())
		if value != "" {
			answerStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)
			return tea.NewView(fmt.Sprintf("%s %s %s", questionMark, promptText, answerStyle.Render(value)))
		}
		return tea.NewView("")
	}

	var b strings.Builder

	// Show help text above the prompt when toggled
	if m.showHelp && m.help != "" {
		helpStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)
		b.WriteString(fmt.Sprintf("%s %s\n", questionMark, helpStyle.Render(m.help)))
	}

	// Main prompt line: ? Prompt <input> [? for help]
	helpHint := ""
	if m.help != "" && !m.showHelp {
		helpHint = lipgloss.NewStyle().Foreground(styles.SubtleColor).Render(" [? for help]")
	}

	b.WriteString(fmt.Sprintf("%s %s %s%s", questionMark, promptText, m.input.View(), helpHint))

	if m.err != "" {
		errStyle := lipgloss.NewStyle().Foreground(styles.ErrorColor)
		b.WriteString(fmt.Sprintf("\n  %s", errStyle.Render(m.err)))
	}

	return tea.NewView(b.String())
}

// Value returns the entered text
func (m TextInputModel) Value() string {
	return strings.TrimSpace(m.input.Value())
}

// PromptText shows an inline text input prompt and returns the entered value.
func PromptText(prompt, placeholder, help string, required, interactive bool) (string, error) {
	if !interactive {
		return "", fmt.Errorf("interactive terminal required for text input")
	}

	model := NewTextInputModel(prompt, placeholder, help, required)
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return "", err
	}

	m := finalModel.(TextInputModel)
	value := m.Value()
	if required && value == "" {
		return "", fmt.Errorf("input cancelled")
	}

	return value, nil
}

package bubbles

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/bins/cli/internal/ui/teaprogram"

	"github.com/nuonco/nuon/pkg/cli/styles"
)

type ConfirmModel struct {
	prompt   string
	result   bool
	selected bool
	quitting bool
	choice   int
}

func NewConfirmModel(prompt string) ConfirmModel {
	return ConfirmModel{
		prompt: prompt,
		choice: 0,
	}
}

func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "left", "shift+tab":
			if m.choice > 0 {
				m.choice--
			}

		case "right", "tab":
			if m.choice < 1 {
				m.choice++
			}

		case "enter", "space":
			m.result = m.choice == 0
			m.selected = true
			m.quitting = true
			return m, tea.Quit

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

func (m ConfirmModel) View() tea.View {
	if m.quitting {
		if m.selected {
			resultText := "No"
			if m.result {
				resultText = "Yes"
			}
			return tea.NewView(SuccessStyle.Render(fmt.Sprintf("✓ %s", resultText)))
		}
		return tea.NewView("")
	}

	promptStyle := lipgloss.NewStyle().
		Foreground(styles.ErrorColor).
		Bold(true).
		Margin(1, 0)

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
		Foreground(styles.SubtleColor).
		Italic(true).
		Margin(1, 0).
		Render("Use arrow keys or tab to navigate • Enter to confirm • y/n for shortcuts")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		promptStyle.Render(m.prompt),
		buttons,
		help,
	)

	return tea.NewView(BorderStyle.Render(content))
}

func (m ConfirmModel) Result() bool {
	return m.result
}

func (m ConfirmModel) Selected() bool {
	return m.selected
}

func Confirm(prompt string, interactive bool) (bool, error) {
	if !interactive {
		return false, fmt.Errorf("interactive terminal required for confirmation; use --yes flag to auto-approve")
	}

	model := NewConfirmModel(prompt)

	program := teaprogram.NewProgram(model)
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

func PromptWithAutoApprove(autoApprove, interactive bool, msg string, vars ...interface{}) error {
	if autoApprove || !interactive {
		return nil
	}

	prompt := fmt.Sprintf(msg, vars...)
	result, err := Confirm(prompt, interactive)
	if err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("operation cancelled")
	}

	return nil
}

func ConfirmWithDefault(prompt string, defaultYes, interactive bool) (bool, error) {
	if !interactive {
		return false, fmt.Errorf("interactive terminal required for confirmation; use --yes flag to auto-approve")
	}

	model := NewConfirmModel(prompt)
	if !defaultYes {
		model.choice = 1
	}

	program := teaprogram.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return defaultYes, err
	}

	confirmModel := finalModel.(ConfirmModel)
	if !confirmModel.Selected() {
		return defaultYes, nil
	}

	return confirmModel.Result(), nil
}

func EvaluationConfirm(prompt string, interactive bool) (bool, error) {
	evaluationPrompt := fmt.Sprintf("🚀 %s\n\n💡 This is part of your Nuon evaluation experience.", prompt)
	return Confirm(evaluationPrompt, interactive)
}

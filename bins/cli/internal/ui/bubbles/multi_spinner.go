package bubbles

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/bins/cli/internal/ui/teaprogram"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

type SpinnerState struct {
	id        string
	message   string
	spinner   spinner.Model
	completed bool
	success   bool
	finalMsg  string
}

type MultiSpinnerModel struct {
	spinners map[string]*SpinnerState
	order    []string
	quitting bool
	width    int
}

type SpinnerUpdateMsg struct {
	ID      string
	Message string
}

type SpinnerCompleteMsg struct {
	ID       string
	Success  bool
	FinalMsg string
}

type FinalRenderMsg struct{}

type RenderCompleteMsg struct{}

func NewMultiSpinner() MultiSpinnerModel {
	return MultiSpinnerModel{
		spinners: make(map[string]*SpinnerState),
		order:    make([]string, 0),
		width:    80,
	}
}

func (m *MultiSpinnerModel) AddSpinner(id, message string) {
	if _, exists := m.spinners[id]; exists {
		return
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.PrimaryColor)

	state := &SpinnerState{
		id:      id,
		message: message,
		spinner: s,
	}

	m.spinners[id] = state
	m.order = append(m.order, id)
}

func (m MultiSpinnerModel) Init() tea.Cmd {
	if len(m.spinners) == 0 {
		return nil
	}

	var cmds []tea.Cmd
	for _, state := range m.spinners {
		if !state.completed {
			cmds = append(cmds, state.spinner.Tick)
		}
	}

	return tea.Batch(cmds...)
}

func (m MultiSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "esc" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case tea.QuitMsg:
		m.quitting = true
		return m, tea.Quit

	case SpinnerUpdateMsg:
		if state, exists := m.spinners[msg.ID]; exists {
			state.message = msg.Message
		}
		return m, nil

	case SpinnerCompleteMsg:
		if state, exists := m.spinners[msg.ID]; exists {
			state.completed = true
			state.success = msg.Success
			state.finalMsg = msg.FinalMsg
		}
		return m, nil

	case FinalRenderMsg:
		return m, func() tea.Msg { return RenderCompleteMsg{} }

	case RenderCompleteMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	default:
		var cmds []tea.Cmd
		for _, state := range m.spinners {
			if !state.completed {
				var cmd tea.Cmd
				state.spinner, cmd = state.spinner.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}

		return m, nil
	}
}

func (m MultiSpinnerModel) View() tea.View {
	if len(m.spinners) == 0 {
		return tea.NewView("")
	}

	var lines []string

	for _, id := range m.order {
		state := m.spinners[id]
		line := m.renderSpinnerLine(state)
		lines = append(lines, line)
	}
	lines = append(lines, "")

	return tea.NewView(strings.Join(lines, "\n"))
}

func (m MultiSpinnerModel) renderSpinnerLine(state *SpinnerState) string {
	if state.completed {
		var icon string
		var style lipgloss.Style

		if state.success {
			icon = "✓"
			style = lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
		} else {
			icon = "✗"
			style = lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true)
		}

		message := state.finalMsg
		if message == "" {
			message = state.message
		}

		return style.Render(fmt.Sprintf("%s %s", icon, message))
	}

	spinnerView := state.spinner.View()
	message := state.message

	maxWidth := m.width - len(spinnerView) - 3
	if maxWidth > 0 && len(message) > maxWidth {
		message = message[:maxWidth-3] + "..."
	}

	return fmt.Sprintf("%s %s", spinnerView, message)
}

func (m MultiSpinnerModel) AllCompleted() bool {
	for _, state := range m.spinners {
		if !state.completed {
			return false
		}
	}
	return true
}

func (m MultiSpinnerModel) HasErrors() bool {
	for _, state := range m.spinners {
		if state.completed && !state.success {
			return true
		}
	}
	return false
}

type MultiSpinnerView struct {
	model       MultiSpinnerModel
	program     *tea.Program
	done        chan struct{}
	started     bool
	interactive bool
}

func NewMultiSpinnerView(interactive bool) *MultiSpinnerView {
	return &MultiSpinnerView{
		model:       NewMultiSpinner(),
		done:        make(chan struct{}),
		interactive: interactive,
	}
}

func (v *MultiSpinnerView) Start() {
	if v.started {
		return
	}

	if len(v.model.spinners) == 0 {
		return
	}

	v.started = true

	if !v.interactive {
		return
	}

	v.program = teaprogram.NewProgram(v.model)

	go func() {
		defer close(v.done)
		if _, err := v.program.Run(); err != nil {
		}
	}()

	time.Sleep(100 * time.Millisecond)
}

func (v *MultiSpinnerView) AddSpinner(id, message string) {
	v.model.AddSpinner(id, message)

	if !v.interactive {
		fmt.Fprintf(agentmode.HumanWriter(), "[%s] %s\n", id, message)
		return
	}

	if v.program != nil {
		v.program.Send(SpinnerUpdateMsg{ID: id, Message: message})
	}
}

func (v *MultiSpinnerView) UpdateSpinner(id, message string) {
	if !v.interactive {
		return
	}

	if v.program == nil {
		return
	}

	v.program.Send(SpinnerUpdateMsg{
		ID:      id,
		Message: message,
	})
}

func (v *MultiSpinnerView) CompleteSpinner(id string, success bool, finalMsg string) {
	if !v.interactive {
		icon := "✓"
		if !success {
			icon = "✗"
		}
		msg := finalMsg
		if msg == "" {
			if state, exists := v.model.spinners[id]; exists {
				msg = state.message
			}
		}
		fmt.Fprintf(agentmode.HumanWriter(), "[%s] %s %s\n", id, icon, msg)
		return
	}

	if v.program == nil {
		return
	}

	v.program.Send(SpinnerCompleteMsg{
		ID:       id,
		Success:  success,
		FinalMsg: finalMsg,
	})
}

func (v *MultiSpinnerView) Stop() {
	if !v.interactive {
		v.started = false
		return
	}

	if v.program == nil {
		return
	}

	v.program.Quit()

	select {
	case <-v.done:
	case <-time.After(2 * time.Second):
		v.program.Kill()
	}

	v.program = nil
	v.started = false
}

func (v *MultiSpinnerView) Wait() {
	if !v.interactive {
		return
	}
	<-v.done
}

func (v *MultiSpinnerView) AllCompleted() bool {
	return v.model.AllCompleted()
}

func (v *MultiSpinnerView) HasErrors() bool {
	return v.model.HasErrors()
}

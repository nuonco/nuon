package bubbles

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// SpinnerState represents the state of an individual spinner
type SpinnerState struct {
	id        string
	message   string
	spinner   spinner.Model
	completed bool
	success   bool
	finalMsg  string
}

// MultiSpinnerModel manages multiple concurrent spinners
type MultiSpinnerModel struct {
	spinners map[string]*SpinnerState
	order    []string // Maintain display order
	quitting bool
	width    int
}

// SpinnerUpdateMsg is used to update spinner messages
type SpinnerUpdateMsg struct {
	ID      string
	Message string
}

// SpinnerCompleteMsg is used to mark spinners as completed
type SpinnerCompleteMsg struct {
	ID       string
	Success  bool
	FinalMsg string
}

// FinalRenderMsg is used to trigger a final render before quitting
type FinalRenderMsg struct{}

// RenderCompleteMsg is used to tell bubbletea that the final render is complete
// TODO(ja): lol wtf
type RenderCompleteMsg struct{}

// NewMultiSpinner creates a new multi-spinner model
func NewMultiSpinner() MultiSpinnerModel {
	return MultiSpinnerModel{
		spinners: make(map[string]*SpinnerState),
		order:    make([]string, 0),
		width:    80,
	}
}

// AddSpinner adds a new spinner to the collection
func (m *MultiSpinnerModel) AddSpinner(id, message string) {
	if _, exists := m.spinners[id]; exists {
		return // Don't add duplicates
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

// Init initializes all spinners
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

// Update handles messages for all spinners
func (m MultiSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
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
		// Trigger one final render, then quit
		return m, func() tea.Msg { return RenderCompleteMsg{} }

	case RenderCompleteMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	default:
		// Update all active spinners
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

// View renders all spinners
func (m MultiSpinnerModel) View() string {
	if len(m.spinners) == 0 {
		return ""
	}

	var lines []string

	for _, id := range m.order {
		state := m.spinners[id]
		// Only show active (not completed) spinners in the live view
		// Completed spinners will be printed manually after program stops
		if !state.completed {
			line := m.renderSpinnerLine(state)
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// renderSpinnerLine renders a single spinner line
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

	// Active spinner
	spinnerView := state.spinner.View()
	message := state.message

	// Truncate message if too long
	maxWidth := m.width - len(spinnerView) - 3
	if maxWidth > 0 && len(message) > maxWidth {
		message = message[:maxWidth-3] + "..."
	}

	return fmt.Sprintf("%s %s", spinnerView, message)
}

// AllCompleted checks if all spinners are completed
func (m MultiSpinnerModel) AllCompleted() bool {
	for _, state := range m.spinners {
		if !state.completed {
			return false
		}
	}
	return true
}

// HasErrors checks if any spinners completed with errors
func (m MultiSpinnerModel) HasErrors() bool {
	for _, state := range m.spinners {
		if state.completed && !state.success {
			return true
		}
	}
	return false
}

// MultiSpinnerView provides a high-level interface for managing multiple spinners
type MultiSpinnerView struct {
	model   MultiSpinnerModel
	program *tea.Program
	done    chan struct{}
	started bool
}

// NewMultiSpinnerView creates a new multi-spinner view
func NewMultiSpinnerView() *MultiSpinnerView {
	return &MultiSpinnerView{
		model: NewMultiSpinner(),
		done:  make(chan struct{}),
	}
}

// Start begins the multi-spinner display
func (v *MultiSpinnerView) Start() {
	// Only start if we haven't already started
	if v.started {
		return
	}

	// Don't start if there are no spinners to display
	if len(v.model.spinners) == 0 {
		return
	}

	v.started = true
	v.program = tea.NewProgram(v.model)

	// Run in background
	go func() {
		defer close(v.done)
		if _, err := v.program.Run(); err != nil {
			// Handle error if needed
		}
	}()

	// Give the program time to initialize
	time.Sleep(100 * time.Millisecond)
}

// AddSpinner adds a new spinner
func (v *MultiSpinnerView) AddSpinner(id, message string) {
	v.model.AddSpinner(id, message)

	// If program is running, send update message
	if v.program != nil {
		v.program.Send(SpinnerUpdateMsg{ID: id, Message: message})
	}
}

// UpdateSpinner updates a spinner's message
func (v *MultiSpinnerView) UpdateSpinner(id, message string) {
	if v.program == nil {
		return
	}

	v.program.Send(SpinnerUpdateMsg{
		ID:      id,
		Message: message,
	})
}

// CompleteSpinner marks a spinner as completed
func (v *MultiSpinnerView) CompleteSpinner(id string, success bool, finalMsg string) {
	if v.program == nil {
		return
	}

	v.program.Send(SpinnerCompleteMsg{
		ID:       id,
		Success:  success,
		FinalMsg: finalMsg,
	})
}

// Stop stops the multi-spinner display
func (v *MultiSpinnerView) Stop() {
	if v.program == nil {
		return
	}

	// Collect completed states before stopping program
	var completedLines []string
	for _, id := range v.model.order {
		state := v.model.spinners[id]
		if state.completed {
			line := v.model.renderSpinnerLine(state)
			completedLines = append(completedLines, line)
		}
	}

	v.program.Quit()

	// Wait for completion with timeout
	select {
	case <-v.done:
		// Clean shutdown
	case <-time.After(2 * time.Second):
		// Force quit if it takes too long
		v.program.Kill()
	}

	// Print the completed states to ensure they persist
	for _, line := range completedLines {
		fmt.Println(line)
	}

	// Reset state
	v.program = nil
	v.started = false
}

// Wait waits for all spinners to complete
func (v *MultiSpinnerView) Wait() {
	<-v.done
}

// AllCompleted checks if all spinners are completed
func (v *MultiSpinnerView) AllCompleted() bool {
	return v.model.AllCompleted()
}

// HasErrors checks if any spinners completed with errors
func (v *MultiSpinnerView) HasErrors() bool {
	return v.model.HasErrors()
}

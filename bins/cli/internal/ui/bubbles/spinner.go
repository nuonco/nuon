package bubbles

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
)

// SpinnerModel represents a spinner with status and message handling
type SpinnerModel struct {
	spinner  spinner.Model
	message  string
	prevText string
	result   *SpinnerResult
	err      error
	quitting bool
	json     bool
	width    int
}

// SpinnerResult represents the final state of a spinner operation
type SpinnerResult struct {
	Success bool
	Message string
	Error   error
}

// SpinnerMsg is used to update the spinner message
type SpinnerMsg string

// SpinnerResultMsg is used to set the final result
type SpinnerResultMsg SpinnerResult

// NewSpinnerModel creates a new spinner model
func NewSpinnerModel(json bool) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.PrimaryColor)

	return SpinnerModel{
		spinner: s,
		json:    json,
		width:   30, // Default width for consistent display
	}
}

// Init initializes the spinner model
func (m SpinnerModel) Init() tea.Cmd {
	if m.json {
		return nil
	}
	return m.spinner.Tick
}

// Update handles messages for the spinner model
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case SpinnerMsg:
		m.message = string(msg)
		return m, nil

	case SpinnerResultMsg:
		m.result = &SpinnerResult{
			Success: msg.Success,
			Message: msg.Message,
			Error:   msg.Error,
		}
		m.quitting = true
		return m, tea.Quit

	default:
		if !m.json && m.result == nil {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}
}

// View renders the spinner
func (m SpinnerModel) View() string {
	if m.json {
		return ""
	}

	if m.result != nil {
		return m.renderResult()
	}

	if m.quitting && m.result == nil {
		return ""
	}

	// Format message with consistent width
	formattedMsg := m.formatText(m.message)
	return fmt.Sprintf("%s %s", m.spinner.View(), formattedMsg)
}

// renderResult renders the final result (success or failure)
func (m SpinnerModel) renderResult() string {
	if m.result.Success {
		style := lipgloss.NewStyle().Foreground(SuccessColor).Bold(true)
		return style.Render(fmt.Sprintf("✓ %s", m.formatText(m.result.Message)))
	}

	// Handle error display
	if m.result.Error != nil {
		errorMsg := m.formatErrorMessage(m.result.Error)
		style := lipgloss.NewStyle().Foreground(ErrorColor).Bold(true)
		return style.Render(fmt.Sprintf("✗ %s", m.formatText(errorMsg)))
	}

	return ""
}

// formatText ensures consistent text width to prevent tearing
func (m SpinnerModel) formatText(text string) string {
	if len(text) > m.width {
		return text[:m.width-3] + "..."
	}

	// Pad with spaces to maintain consistent width
	for len(text) < len(m.prevText) {
		text += " "
	}

	m.prevText = text
	return text
}

// formatErrorMessage formats error messages similar to the original pterm logic
func (m SpinnerModel) formatErrorMessage(err error) string {
	if hints := errors.FlattenHints(err); hints != "" {
		return hints
	}

	if userErr, ok := nuon.ToUserError(err); ok {
		return userErr.Description
	}

	if nuon.IsServerError(err) {
		return "Oops, we have experienced a server error. Please try again in a few minutes."
	}

	return err.Error()
}

// SpinnerView provides a high-level interface similar to the original spinner_view
type SpinnerView struct {
	json    bool
	program *tea.Program
	model   SpinnerModel
}

// NewSpinnerView creates a new spinner view
func NewSpinnerView(json bool) *SpinnerView {
	return &SpinnerView{
		json: json,
	}
}

// Start begins the spinner with the given message
func (v *SpinnerView) Start(text string) {
	if v.json {
		return
	}

	v.model = NewSpinnerModel(v.json)
	v.model.message = text

	v.program = tea.NewProgram(v.model)

	// Run in background
	go func() {
		if _, err := v.program.Run(); err != nil {
			// Handle error if needed
		}
	}()
}

// Update changes the spinner message
func (v *SpinnerView) Update(text string) {
	if v.json || v.program == nil {
		return
	}

	v.program.Send(SpinnerMsg(text))
}

// Success completes the spinner with a success message
func (v *SpinnerView) Success(text string) {
	if v.json {
		// Handle JSON output if needed
		fmt.Println(text)
		return
	}

	if v.program == nil {
		return
	}

	v.program.Send(SpinnerResultMsg{
		Success: true,
		Message: text,
	})

	// Give the UI time to render the result
	time.Sleep(100 * time.Millisecond)
}

// Fail completes the spinner with an error message
func (v *SpinnerView) Fail(err error) {
	if v.json {
		// Handle JSON error output
		fmt.Printf(`{"error": "%s"}`, err.Error())
		return
	}

	if v.program == nil {
		return
	}

	v.program.Send(SpinnerResultMsg{
		Success: false,
		Error:   err,
	})

	// Give the UI time to render the result
	time.Sleep(100 * time.Millisecond)
}

// RunSpinnerWithContext runs a spinner for the duration of a context operation
func RunSpinnerWithContext(ctx context.Context, message string, operation func(ctx context.Context) error, json bool) error {
	if json {
		return operation(ctx)
	}

	spinnerView := NewSpinnerView(json)
	spinnerView.Start(message)

	err := operation(ctx)

	if err != nil {
		spinnerView.Fail(err)
	} else {
		spinnerView.Success(message + " completed")
	}

	return err
}


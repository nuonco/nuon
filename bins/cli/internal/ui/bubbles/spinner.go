package bubbles

import (
	"context"
	"fmt"
	"sync"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go"
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
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
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
func (m SpinnerModel) View() tea.View {
	if m.json {
		return tea.NewView("")
	}

	if m.result != nil {
		return tea.NewView(m.renderResult())
	}

	if m.quitting && m.result == nil {
		return tea.NewView("")
	}

	// Format message with consistent width
	formattedMsg := m.formatText(m.message)
	return tea.NewView(fmt.Sprintf("%s %s", m.spinner.View(), formattedMsg))
}

// renderResult renders the final result (success or failure)
func (m SpinnerModel) renderResult() string {
	if m.result.Success {
		style := lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
		return style.Render(fmt.Sprintf("✓ %s", m.result.Message))
	}

	// Handle error display
	if m.result.Error != nil {
		errorMsg := formatErrorMessage(m.result.Error)
		style := lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true)
		return style.Render(fmt.Sprintf("✗ %s", errorMsg))
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
func formatErrorMessage(err error) string {
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
	json        bool
	interactive bool
	program     *tea.Program
	model       SpinnerModel
	wg          sync.WaitGroup
	mu          sync.Mutex
	noTTY       bool
	programErr  error // stores the error from program.Run() if any
	lastUpdate  string
}

// NewSpinnerView creates a new spinner view
func NewSpinnerView(json, interactive bool) *SpinnerView {
	return &SpinnerView{
		json:        json,
		interactive: interactive,
	}
}

// Start begins the spinner with the given message
func (v *SpinnerView) Start(text string) {
	if v.json {
		return
	}

	v.lastUpdate = text

	if !v.interactive {
		fmt.Println(text)
		return
	}

	v.model = NewSpinnerModel(v.json)
	v.model.message = text

	v.program = tea.NewProgram(v.model)

	// Run in background
	v.wg.Add(1)
	go func() {
		defer v.wg.Done()
		if _, err := v.program.Run(); err != nil {
			// If bubbletea fails (no TTY), fall back to simple mode
			v.mu.Lock()
			v.noTTY = true
			v.programErr = err
			v.mu.Unlock()
		}
	}()
}

// Update changes the spinner message
func (v *SpinnerView) Update(text string) {
	if v.json {
		return
	}

	if !v.interactive {
		v.lastUpdate = text
		return
	}

	if v.program == nil {
		return
	}

	v.program.Send(SpinnerMsg(text))
}

// Success completes the spinner with a success message
func (v *SpinnerView) Success(text string) {
	if v.json {
		fmt.Println(text)
		return
	}

	if !v.interactive {
		fmt.Printf("✓ %s\n", text)
		return
	}

	if v.program == nil {
		return
	}

	v.program.Send(SpinnerResultMsg{
		Success: true,
		Message: text,
	})

	// Wait for the program to finish
	v.wg.Wait()

	// Check if program failed (no TTY) after waiting for it to complete.
	// If so, the program didn't render the success, so print it directly.
	v.mu.Lock()
	noTTY := v.noTTY
	v.mu.Unlock()

	if noTTY {
		style := lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
		fmt.Println(style.Render(fmt.Sprintf("✓ %s", text)))
	}
}

// Fail completes the spinner with an error message
func (v *SpinnerView) Fail(err error) {
	if v.json {
		fmt.Printf(`{"error": "%s"}`, err.Error())
		return
	}

	if !v.interactive {
		fmt.Printf("✗ %s\n", formatErrorMessage(err))
		return
	}

	if v.program == nil {
		return
	}

	// Send the result message to the program
	v.program.Send(SpinnerResultMsg{
		Success: false,
		Error:   err,
	})

	// Wait for the program to finish
	v.wg.Wait()

	// Check if program failed (no TTY) after waiting for it to complete.
	// If so, the program didn't render the error, so print it directly.
	v.mu.Lock()
	noTTY := v.noTTY
	v.mu.Unlock()

	if noTTY {
		errorMsg := formatErrorMessage(err)
		style := lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true)
		fmt.Println(style.Render(fmt.Sprintf("✗ %s", errorMsg)))
	}
}

// RunSpinnerWithContext runs a spinner for the duration of a context operation
func RunSpinnerWithContext(ctx context.Context, message string, operation func(ctx context.Context) error, json, interactive bool) error {
	if json {
		return operation(ctx)
	}

	spinnerView := NewSpinnerView(json, interactive)
	spinnerView.Start(message)

	err := operation(ctx)

	if err != nil {
		spinnerView.Fail(err)
	} else {
		spinnerView.Success(message + " completed")
	}

	return err
}

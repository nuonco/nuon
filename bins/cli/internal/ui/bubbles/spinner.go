// Package bubbles owns the CLI's small reusable terminal-UI primitives.
//
// The single-line SpinnerView in this file is intentionally NOT built on
// bubbletea. The bubbletea v2 renderer probes for terminal capabilities
// (DEC private modes 2026 "Synchronized Output" and 2027 "Unicode Core")
// the first time it starts. For long-running TUIs that's harmless — the
// terminal's reply arrives while bubbletea is still reading input. For
// short-lived spinners (e.g. `nuon orgs webhooks delete`), the program
// quits via tea.Quit a few hundred milliseconds later when the API call
// returns, well before the terminal flushes its reply. The reply then
// leaks to the user's shell and shows up at the next prompt as raw
// bytes like `^[[?2026;2$y^[[?2027;2$y`.
//
// The fix is to keep the spinner off bubbletea entirely: render a single
// line directly to stdout with `\r` overwrites and ANSI clear-to-EOL.
// This matches the visual output of the previous spinner.Dot model,
// supports the same Start / Update / Success / Fail surface, and can
// never trigger a terminal-capability probe because we don't construct
// a tea.Program at all.
package bubbles

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

// spinnerFrames matches the visual style of bubbletea's spinner.Dot
// preset so the migration off bubbletea is invisible to users.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const spinnerInterval = 100 * time.Millisecond

// formatErrorMessage formats error messages similar to the original pterm logic.
// Kept as a package-level helper because DeleteView et al. wrap it indirectly
// via SpinnerView.Fail.
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

// SpinnerView is a single-line spinner that renders directly to stdout
// using carriage-return overwrites. It deliberately avoids bubbletea
// because the v2 renderer probes for terminal capabilities at startup
// and the response leaks to the user's shell on short-lived programs.
type SpinnerView struct {
	json        bool
	interactive bool
	out         io.Writer

	mu      sync.Mutex
	running bool
	msg     string
	done    chan struct{}
	wg      sync.WaitGroup
}

// NewSpinnerView creates a new spinner view.
func NewSpinnerView(json, interactive bool) *SpinnerView {
	return &SpinnerView{
		json:        json,
		interactive: interactive,
		out:         os.Stdout,
	}
}

// Start begins the spinner with the given message.
func (v *SpinnerView) Start(text string) {
	if v.json {
		return
	}

	if !v.interactive {
		fmt.Fprintln(v.out, text)
		return
	}

	v.mu.Lock()
	if v.running {
		v.msg = text
		v.mu.Unlock()
		return
	}
	v.running = true
	v.msg = text
	v.done = make(chan struct{})
	v.mu.Unlock()

	v.wg.Add(1)
	go v.run()
}

func (v *SpinnerView) run() {
	defer v.wg.Done()

	style := lipgloss.NewStyle().Foreground(styles.PrimaryColor)
	ticker := time.NewTicker(spinnerInterval)
	defer ticker.Stop()

	frame := 0
	for {
		v.mu.Lock()
		msg := v.msg
		v.mu.Unlock()
		// \r returns to col 0; \033[2K clears the whole line so a
		// shorter message after Update doesn't leave trailing chars.
		fmt.Fprintf(v.out, "\r\033[2K%s %s", style.Render(spinnerFrames[frame]), msg)
		frame = (frame + 1) % len(spinnerFrames)

		select {
		case <-v.done:
			// Clear the spinner line before the caller prints the
			// terminal success / failure result.
			fmt.Fprint(v.out, "\r\033[2K")
			return
		case <-ticker.C:
		}
	}
}

// Update changes the spinner message.
func (v *SpinnerView) Update(text string) {
	if v.json {
		return
	}
	if !v.interactive {
		// Mirror the bubbletea version's no-print update behaviour:
		// the next Start/Success/Fail will surface the new message.
		v.mu.Lock()
		v.msg = text
		v.mu.Unlock()
		return
	}

	v.mu.Lock()
	v.msg = text
	v.mu.Unlock()
}

// stop signals the render goroutine to exit and waits for it.
func (v *SpinnerView) stop() {
	v.mu.Lock()
	if !v.running {
		v.mu.Unlock()
		return
	}
	v.running = false
	close(v.done)
	v.mu.Unlock()
	v.wg.Wait()
}

// Success completes the spinner with a success message.
func (v *SpinnerView) Success(text string) {
	if v.json {
		fmt.Fprintln(v.out, text)
		return
	}

	if !v.interactive {
		fmt.Fprintf(v.out, "✓ %s\n", text)
		return
	}

	v.stop()
	style := lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
	fmt.Fprintln(v.out, style.Render(fmt.Sprintf("✓ %s", text)))
}

// Fail completes the spinner with an error message.
func (v *SpinnerView) Fail(err error) {
	if v.json {
		fmt.Fprintf(v.out, `{"error": "%s"}`+"\n", err.Error())
		return
	}

	errorMsg := formatErrorMessage(err)
	if !v.interactive {
		fmt.Fprintf(v.out, "✗ %s\n", errorMsg)
		return
	}

	v.stop()
	style := lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true)
	fmt.Fprintln(v.out, style.Render(fmt.Sprintf("✗ %s", errorMsg)))
}

// RunSpinnerWithContext runs a spinner for the duration of a context operation.
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

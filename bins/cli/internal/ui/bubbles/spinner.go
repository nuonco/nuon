// why: Package bubbles owns the CLI's small reusable terminal-UI primitives.
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
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/cockroachdb/errors"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const spinnerInterval = 100 * time.Millisecond

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

func NewSpinnerView(json, interactive bool) *SpinnerView {
	return &SpinnerView{
		json:        json,
		interactive: interactive,
		out:         agentmode.HumanWriter(),
	}
}

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
		fmt.Fprintf(v.out, "\r\033[2K%s %s", style.Render(spinnerFrames[frame]), msg)
		frame = (frame + 1) % len(spinnerFrames)

		select {
		case <-v.done:
			fmt.Fprint(v.out, "\r\033[2K")
			return
		case <-ticker.C:
		}
	}
}

func (v *SpinnerView) Update(text string) {
	if v.json {
		return
	}
	if !v.interactive {
		v.mu.Lock()
		v.msg = text
		v.mu.Unlock()
		return
	}

	v.mu.Lock()
	v.msg = text
	v.mu.Unlock()
}

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

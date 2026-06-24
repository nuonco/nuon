// Package tui renders the full-screen interactive pre-provision walkthrough
// for stack-cli using bubbletea v2.
//
// Six steps in order: Auth (review) → Inputs & Secrets → Network (review) →
// Roles → Method → Confirm. The user's stack.Config is mutated in place;
// the Installer reads the same pointer via preCreatedRun.Config.
package tui

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/sdks/stack"
)

// ErrAborted is returned when the user cancels the walkthrough.
var ErrAborted = errors.New("aborted by user")

// Run drives the walkthrough in alt-screen mode. cfg is mutated in place.
// kind picks the operation-specific header label (Provision / Reprovision /
// Deprovision Stack).
func Run(ctx context.Context, kind stack.Kind, cfg *stack.Config) error {
	m := newAppModel(ctx, kind, cfg)

	// Read keystrokes from the controlling terminal, not stdin: the dashboard
	// runs the wizard via `curl … | sh`, which leaves stdin attached to the
	// (already-consumed) pipe. /dev/tty gives us real interactive input.
	opts := []tea.ProgramOption{}
	if tty, err := os.Open("/dev/tty"); err == nil {
		defer tty.Close()
		opts = append(opts, tea.WithInput(tty))
	}

	p := tea.NewProgram(m, opts...)
	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("tui run: %w", err)
	}
	fm, ok := final.(appModel)
	if !ok {
		return fmt.Errorf("tui exited with unexpected model")
	}
	if fm.err != nil {
		return fm.err
	}
	if !fm.confirmed {
		return ErrAborted
	}
	return nil
}

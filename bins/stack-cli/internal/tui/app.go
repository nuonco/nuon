package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/sdks/stack"
)

// step identifies which slot in the walkthrough is active.
type step int

const (
	stepIntro step = iota
	stepMethod
	stepAuth
	stepInputs
	stepNetwork
	stepRoles
	stepConfirm
)

// stepModel is what each step must satisfy. Steps own their own focus and
// render their own Previous/Next buttons; the parent provides chrome and
// listens for nav messages.
type stepModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (stepModel, tea.Cmd)
	View(width, height int) string
	Help() string
	// CanAdvance gates nextStepMsg; when false, the reason surfaces in the
	// footer instead of advancing.
	CanAdvance() (bool, string)
}

// nextStepMsg / prevStepMsg / finishMsg are emitted by steps when their
// internal buttons are activated. The app translates these into step
// transitions.
type (
	nextStepMsg struct{}
	prevStepMsg struct{}
	finishMsg   struct{}
	abortMsg    struct{}
)

type appModel struct {
	ctx  context.Context
	cfg  *stack.Config
	kind stack.Kind

	current step
	steps   map[step]stepModel

	width, height int
	flash         string // transient status text shown in footer

	confirmed bool
	err       error
}

func newAppModel(ctx context.Context, kind stack.Kind, cfg *stack.Config) appModel {
	return appModel{
		ctx:     ctx,
		cfg:     cfg,
		kind:    kind,
		current: stepIntro,
		steps: map[step]stepModel{
			stepIntro:   newIntroStep(kind),
			stepMethod:  newMethodStep(cfg),
			stepAuth:    newAuthStep(ctx, cfg),
			stepInputs:  newInputsStep(cfg),
			stepNetwork: newNetworkStep(cfg),
			stepRoles:   newRolesStep(cfg),
			stepConfirm: newConfirmStep(cfg),
		},
	}
}

func (m appModel) Init() tea.Cmd {
	return m.steps[m.current].Init()
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if s := msg.String(); s == "ctrl+c" || s == "esc" {
			return m, tea.Quit
		}

	case nextStepMsg:
		return m.advance()

	case prevStepMsg:
		return m.goBack()

	case abortMsg:
		return m, tea.Quit

	case finishMsg:
		m.confirmed = true
		return m, tea.Quit
	}

	updated, cmd := m.steps[m.current].Update(msg)
	m.steps[m.current] = updated
	return m, cmd
}

func (m appModel) goBack() (tea.Model, tea.Cmd) {
	if m.current == stepIntro {
		return m, nil
	}
	m.current--
	m.flash = ""
	return m, m.steps[m.current].Init()
}

func (m appModel) advance() (tea.Model, tea.Cmd) {
	ok, why := m.steps[m.current].CanAdvance()
	if !ok {
		m.flash = why
		return m, nil
	}
	m.flash = ""
	if m.current == stepConfirm {
		m.confirmed = true
		return m, tea.Quit
	}
	m.current++
	// When entering Provision, copy the resolved AWS account ID from Auth so
	// the summary shows it without a second STS call.
	if m.current == stepConfirm {
		if auth, ok := m.steps[stepAuth].(*authStep); ok {
			if confirm, ok := m.steps[stepConfirm].(*confirmStep); ok {
				confirm.accountID = auth.account
			}
		}
	}
	return m, m.steps[m.current].Init()
}

func (m appModel) View() tea.View {
	w, h := m.width, m.height
	if w == 0 || h == 0 {
		v := tea.NewView("loading…")
		v.AltScreen = true
		return v
	}
	if w < minWidth || h < minHeight {
		msg := fmt.Sprintf("Terminal too small (need %dx%d, have %dx%d) — resize or use --non-interactive.", minWidth, minHeight, w, h)
		v := tea.NewView(errStyle.Render(msg))
		v.AltScreen = true
		return v
	}

	// outerStyle adds (margin 1×2) + (border 1) + (padding 0×1) on each side.
	// Horizontal subtract: 2+1+1+1+1+2 = 8. Vertical subtract: 1+1+1+1 = 4.
	innerW := w - 8
	innerH := h - 4

	header := m.renderHeader(innerW)
	stepper := m.renderStepper(innerW)
	footer := m.renderFooter(innerW)

	// Chrome: header(1) + rule(1) + stepper(1) + rule(1) + body + rule(1) + footer(1) = 6 + body.
	bodyH := innerH - 6
	if bodyH < 4 {
		bodyH = 4
	}

	// bodyPadStyle has Padding(1, 2): 1 row top + 1 row bottom. The step's
	// content area is therefore bodyH-2 rows; tell the step that up front so
	// it renders to the right size, and clip defensively.
	stepH := bodyH - 2
	if stepH < 1 {
		stepH = 1
	}
	bodyContent := m.steps[m.current].View(innerW, stepH)
	bodyContent = clipLinesExact(bodyContent, stepH)
	body := bodyPadStyle.
		Width(innerW).
		Height(bodyH).
		Render(bodyContent)

	rule := ruleStyle.Render(strings.Repeat("─", innerW))

	inner := lipgloss.JoinVertical(lipgloss.Left,
		header,
		rule,
		stepper,
		rule,
		body,
		rule,
		footer,
	)
	v := tea.NewView(outerStyle.Render(inner))
	v.AltScreen = true
	return v
}

func (m appModel) renderHeader(w int) string {
	left := headerStyle.Render(operationTitle(m.kind))
	right := headerMetaStyle.Render(fmt.Sprintf("region: %s", m.cfg.AWSRegion))
	if auth, ok := m.steps[stepAuth].(*authStep); ok && auth.account != "" {
		right = headerMetaStyle.Render(fmt.Sprintf("region: %s · account: %s",
			m.cfg.AWSRegion, auth.account))
	}
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m appModel) renderStepper(w int) string {
	labels := []string{"Method", "Auth", "Inputs", "Network", "Roles", "Provision"}
	parts := make([]string, 0, len(labels)*2)
	for i, lbl := range labels {
		// labels[0] corresponds to stepMethod (step value 1); the intro step
		// (value 0) is intentionally not in the stepper.
		stepValue := step(i + 1)
		var marker, text string
		switch {
		case stepValue < m.current:
			marker = stepDone.Render("✓")
			text = stepDone.Render(lbl)
		case stepValue == m.current:
			marker = stepActive.Render(fmt.Sprintf("%d", i+1))
			text = stepActive.Render(lbl)
		default:
			marker = stepPending.Render(fmt.Sprintf("%d", i+1))
			text = stepPending.Render(lbl)
		}
		parts = append(parts, marker+" "+text)
	}
	joined := strings.Join(parts, stepSep.Render("  ▸  "))
	return lipgloss.NewStyle().Padding(0, 1).Render(joined)
}

func (m appModel) renderFooter(w int) string {
	help := m.steps[m.current].Help()
	common := footerKey.Render("enter") + footerStyle.Render(" select  ") +
		footerKey.Render("esc") + footerStyle.Render(" cancel")
	left := footerStyle.Render(help)
	if m.flash != "" {
		left = errStyle.Render(m.flash)
	}
	gap := w - lipgloss.Width(left) - lipgloss.Width(common) - 2
	if gap < 1 {
		gap = 1
	}
	return " " + left + strings.Repeat(" ", gap) + common + " "
}

// clipLinesExact caps s to at most n lines so the body never pushes chrome
// off-screen. Steps that have more content than fits are responsible for
// their own scrolling.
func clipLinesExact(s string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}

func operationTitle(k stack.Kind) string {
	switch k {
	case stack.KindReprovision:
		return "Reprovision Stack"
	case stack.KindDeprovision:
		return "Deprovision Stack"
	default:
		return "Provision Stack"
	}
}

func short(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[:8] + "…"
}

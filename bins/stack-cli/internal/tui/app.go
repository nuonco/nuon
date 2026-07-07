package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/sdks/stack"
)

// stepModel is what each step must satisfy. Steps own their own focus and
// render their own Previous/Next buttons; the parent provides chrome, the
// two-column layout, and listens for nav messages. Main renders the left
// column (primary content + buttons); Detail renders the focus-driven right
// column and may return "" when there is nothing to show.
type stepModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (stepModel, tea.Cmd)
	Main(width, height int) string
	Detail(width, height int) string
	Help() string
	// CanAdvance gates nextStepMsg; when false, the reason surfaces in the
	// footer instead of advancing.
	CanAdvance() (bool, string)
}

// stepEntry pairs a step with its stepper label. An empty label means the step
// is not shown in the stepper (the intro screen).
type stepEntry struct {
	label string
	model stepModel
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

	// steps is the ordered walkthrough for this run's cloud; current indexes
	// into it. Index 0 is the intro (label "").
	steps   []stepEntry
	current int

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
		steps:   buildSteps(ctx, kind, cfg),
		current: 0,
	}
}

// buildSteps returns the ordered step set for the run's cloud. Shared steps
// (intro, inputs, confirm) are reused; cloud-specific steps differ.
func buildSteps(ctx context.Context, kind stack.Kind, cfg *stack.Config) []stepEntry {
	switch cfg.Cloud {
	case stack.CloudGCP:
		return gcpSteps(ctx, kind, cfg)
	default:
		return awsSteps(ctx, kind, cfg)
	}
}

func awsSteps(ctx context.Context, kind stack.Kind, cfg *stack.Config) []stepEntry {
	return []stepEntry{
		{"", newIntroStep(kind, stack.CloudAWS)},
		{"Auth", newAuthStep(ctx, cfg)},
		{"Inputs", newInputsStep(cfg)},
		{"Secrets", newSecretsStep(cfg)},
		{"Network", newNetworkStep(cfg)},
		{"Roles", newRolesStep(cfg)},
		{"Provision", newConfirmStep(cfg)},
	}
}

func (m appModel) Init() tea.Cmd {
	return m.steps[m.current].model.Init()
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

	updated, cmd := m.steps[m.current].model.Update(msg)
	m.steps[m.current].model = updated
	return m, cmd
}

func (m appModel) goBack() (tea.Model, tea.Cmd) {
	if m.current == 0 {
		return m, nil
	}
	m.current--
	m.flash = ""
	return m, m.steps[m.current].model.Init()
}

func (m appModel) advance() (tea.Model, tea.Cmd) {
	ok, why := m.steps[m.current].model.CanAdvance()
	if !ok {
		m.flash = why
		return m, nil
	}
	m.flash = ""
	if m.current == len(m.steps)-1 {
		m.confirmed = true
		return m, tea.Quit
	}
	m.current++
	// When entering the confirm step, copy the resolved AWS account ID from the
	// auth step so the summary shows it without a second STS call.
	if m.current == len(m.steps)-1 {
		m.wireConfirm()
	}
	return m, m.steps[m.current].model.Init()
}

// wireConfirm passes data the confirm step wants but doesn't own (the resolved
// AWS account ID from the auth step). No-op for clouds without those steps.
func (m appModel) wireConfirm() {
	var auth *authStep
	var confirm *confirmStep
	for _, e := range m.steps {
		switch s := e.model.(type) {
		case *authStep:
			auth = s
		case *confirmStep:
			confirm = s
		}
	}
	if auth != nil && confirm != nil {
		confirm.accountID = auth.account
	}
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
	cur := m.steps[m.current].model
	bodyContent := twoColumn(innerW, stepH, cur.Main, cur.Detail)
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
	right := headerMetaStyle.Render(m.headerMeta())
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// headerMeta is the cloud-specific location summary shown top-right.
func (m appModel) headerMeta() string {
	switch m.cfg.Cloud {
	case stack.CloudGCP:
		return ""
	default:
		region := ""
		if m.cfg.AWS != nil {
			region = m.cfg.AWS.Region
		}
		for _, e := range m.steps {
			if auth, ok := e.model.(*authStep); ok && auth.account != "" {
				return fmt.Sprintf("region: %s · account: %s", region, auth.account)
			}
		}
		return fmt.Sprintf("region: %s", region)
	}
}

func (m appModel) renderStepper(w int) string {
	// Build labels from the steps that opt into the stepper (non-empty label),
	// numbered sequentially. Their slice index drives done/active/pending.
	parts := make([]string, 0, len(m.steps))
	n := 0
	for i, e := range m.steps {
		if e.label == "" {
			continue
		}
		n++
		var marker, text string
		switch {
		case i < m.current:
			marker = stepDone.Render("✓")
			text = stepDone.Render(e.label)
		case i == m.current:
			marker = stepActive.Render(fmt.Sprintf("%d", n))
			text = stepActive.Render(e.label)
		default:
			marker = stepPending.Render(fmt.Sprintf("%d", n))
			text = stepPending.Render(e.label)
		}
		parts = append(parts, marker+" "+text)
	}
	joined := strings.Join(parts, stepSep.Render("  ▸  "))
	return lipgloss.NewStyle().Padding(0, 1).Render(joined)
}

func (m appModel) renderFooter(w int) string {
	help := m.steps[m.current].model.Help()
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

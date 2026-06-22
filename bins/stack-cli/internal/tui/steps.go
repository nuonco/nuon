package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/nuonco/nuon/sdks/stack"
)

// ─── Intro ──────────────────────────────────────────────────────────────────

type introStep struct {
	kind stack.Kind
}

func newIntroStep(kind stack.Kind) *introStep { return &introStep{kind: kind} }
func (s *introStep) Init() tea.Cmd            { return nil }

func (s *introStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "enter" {
		return s, func() tea.Msg { return nextStepMsg{} }
	}
	return s, nil
}

func (s *introStep) View(w, h int) string {
	verb := "provision"
	switch s.kind {
	case stack.KindReprovision:
		verb = "reprovision"
	case stack.KindDeprovision:
		verb = "deprovision"
	}

	heading := titleStyle.Render("Welcome")

	var body string
	switch s.kind {
	case stack.KindDeprovision:
		body = strings.Join([]string{
			"A " + kvKeyStyle.Render("stack") + " is the set of AWS resources installed in your",
			"account: a VPC and subnets, IAM roles, Secrets Manager",
			"entries, and a runner EC2 instance. Applications run",
			"inside this stack.",
			"",
			kvKeyStyle.Render("Deprovision") + " tears everything down in reverse order, leaving",
			"the AWS account empty of stack-managed resources.",
			"",
			dimStyle.Render("This will delete:"),
			dimStyle.Render("  · the runner EC2 instance and log group"),
			dimStyle.Render("  · all created IAM roles and policies"),
			dimStyle.Render("  · stack secrets in AWS Secrets Manager"),
			dimStyle.Render("  · the VPC, subnets, NAT gateway, and route tables"),
		}, "\n")
	default:
		body = strings.Join([]string{
			"A " + kvKeyStyle.Render("stack") + " is the set of AWS resources installed in your",
			"account so applications can run there: a VPC and subnets,",
			"IAM roles, Secrets Manager entries, and a runner EC2",
			"instance that connects to the control plane.",
			"",
			"This wizard walks you through " + verb + "ing one. You'll:",
			"",
			"  " + stepActive.Render("1.") + " " + kvKeyStyle.Render("Method") + "   — choose how the stack is provisioned",
			"  " + stepActive.Render("2.") + " " + kvKeyStyle.Render("Auth") + "     — verify your AWS credentials and region",
			"  " + stepActive.Render("3.") + " " + kvKeyStyle.Render("Inputs") + "   — fill in install inputs and required secrets",
			"  " + stepActive.Render("4.") + " " + kvKeyStyle.Render("Network") + "  — review the VPC plan",
			"  " + stepActive.Render("5.") + " " + kvKeyStyle.Render("Roles") + "    — pick which IAM roles to create",
			"  " + stepActive.Render("6.") + " " + kvKeyStyle.Render("Provision") + " — confirm the plan and apply it",
			"",
			dimStyle.Render("Nothing is created in your AWS account until you reach the final"),
			dimStyle.Render("step and select Provision."),
		}, "\n")
	}

	next := renderButton(" Get started ▸ ", true, false)
	return heading + "\n\n" + body + "\n\n" + next
}

func (s *introStep) Help() string               { return "" }
func (s *introStep) CanAdvance() (bool, string) { return true, "" }

// ─── Method ───────────────────────────────────────────────────────────────────

type methodOption struct {
	method stack.Method
	label  string
	desc   []string
}

type methodStep struct {
	cfg      *stack.Config
	options  []methodOption
	selected int // index into options that is currently chosen
	cursor   int // 0..len(options)-1 = options, then Previous, then Next
}

func newMethodStep(cfg *stack.Config) *methodStep {
	options := []methodOption{
		{
			method: stack.MethodSDK,
			label:  "AWS SDK",
			desc: []string{
				"Provisions resources directly via the AWS SDK.",
				"No CLI tooling required on this machine — the",
				"stack drives every API call itself.",
			},
		},
		{
			method: stack.MethodTerraform,
			label:  "Terraform module",
			desc: []string{
				"Applies the install-stacks/aws Terraform module.",
				"Terraform is fetched automatically; no local",
				"install required. Phone-home stays in the module.",
			},
		},
	}

	s := &methodStep{cfg: cfg, options: options, cursor: 0}
	s.selected = 0
	for i, o := range options {
		if o.method == cfg.Method {
			s.selected = i
			s.cursor = i
			break
		}
	}
	return s
}

func (s *methodStep) Init() tea.Cmd { return nil }

func (s *methodStep) prevIdx() int { return len(s.options) }
func (s *methodStep) nextIdx() int { return len(s.options) + 1 }

func (s *methodStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	switch k.String() {
	case "down", "j":
		if s.cursor < s.nextIdx() {
			s.cursor++
		}
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "right", "l":
		if s.cursor == s.prevIdx() {
			s.cursor = s.nextIdx()
		}
	case "left", "h":
		if s.cursor == s.nextIdx() {
			s.cursor = s.prevIdx()
		}
	case " ", "x":
		if s.cursor < len(s.options) {
			s.selected = s.cursor
		}
	case "enter":
		switch s.cursor {
		case s.prevIdx():
			return s, func() tea.Msg { return prevStepMsg{} }
		case s.nextIdx():
			s.apply()
			return s, func() tea.Msg { return nextStepMsg{} }
		default:
			s.selected = s.cursor
		}
	}
	return s, nil
}

func (s *methodStep) apply() {
	s.cfg.Method = s.options[s.selected].method
}

func (s *methodStep) View(w, h int) string {
	title := titleStyle.Render("Provisioning method")
	intro := dimStyle.Render("Choose how the stack's AWS resources are created. Both methods")
	intro2 := dimStyle.Render("run self-contained — no CLI tools need to be installed here.")

	var lines []string
	for i, o := range s.options {
		radio := checkboxOff.Render("( )")
		if i == s.selected {
			radio = checkboxOn.Render("(•)")
		}
		focused := i == s.cursor
		marker := "  "
		label := o.label
		if focused {
			marker = focusedStyle.Render("▸ ")
			label = focusedStyle.Render(label)
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", marker, radio, label))
		for _, dl := range o.desc {
			lines = append(lines, "      "+dimStyle.Render(dl))
		}
		lines = append(lines, "")
	}

	prev := renderButton(" ◂ Previous ", s.cursor == s.prevIdx(), false)
	next := renderButton(" Next ▸ ", s.cursor == s.nextIdx(), false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", next)

	body := strings.Join(lines, "\n")
	return title + "\n\n" + intro + "\n" + intro2 + "\n\n" + body + "\n" + buttons
}

func (s *methodStep) Help() string { return "↑↓ move · space select" }

func (s *methodStep) CanAdvance() (bool, string) {
	s.apply()
	return true, ""
}

// ─── Auth ───────────────────────────────────────────────────────────────────

type authStep struct {
	ctx     context.Context
	cfg     *stack.Config
	spinner spinner.Model

	loading bool
	account string
	arn     string
	err     error
}

type authDoneMsg struct {
	account, arn string
	err          error
}

func newAuthStep(ctx context.Context, cfg *stack.Config) *authStep {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &authStep{ctx: ctx, cfg: cfg, spinner: sp, loading: true}
}

func (s *authStep) Init() tea.Cmd {
	if !s.loading && s.err == nil {
		return nil
	}
	s.loading = true
	return tea.Batch(s.spinner.Tick, s.fetch)
}

func (s *authStep) fetch() tea.Msg {
	awsCfg, err := awsconfig.LoadDefaultConfig(s.ctx, awsconfig.WithRegion(s.cfg.AWSRegion))
	if err != nil {
		return authDoneMsg{err: fmt.Errorf("load aws config: %w", err)}
	}
	out, err := sts.NewFromConfig(awsCfg).GetCallerIdentity(s.ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return authDoneMsg{err: fmt.Errorf("sts get-caller-identity: %w", err)}
	}
	return authDoneMsg{account: aws.ToString(out.Account), arn: aws.ToString(out.Arn)}
}

func (s *authStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	if m, ok := msg.(authDoneMsg); ok {
		s.loading = false
		s.account = m.account
		s.arn = m.arn
		s.err = m.err
		return s, nil
	}
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "enter" {
		if !s.loading && s.err == nil {
			return s, func() tea.Msg { return nextStepMsg{} }
		}
		return s, nil
	}
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd
}

func (s *authStep) View(w, h int) string {
	title := titleStyle.Render("Verify AWS credentials")
	var body string
	switch {
	case s.loading:
		body = s.spinner.View() + " " + dimStyle.Render("Calling sts:GetCallerIdentity…")
	case s.err != nil:
		body = errStyle.Render("Error: "+s.err.Error()) + "\n\n" +
			dimStyle.Render("Fix your AWS credentials (env, profile, or SSO) and restart.")
	default:
		rows := []string{
			kvRow("Account", s.account),
			kvRow("ARN", s.arn),
			kvRow("Region", s.cfg.AWSRegion),
		}
		body = strings.Join(rows, "\n") + "\n\n" + renderButton(" Continue ▸ ", true, false)
	}
	return title + "\n\n" + body
}

func (s *authStep) Help() string {
	if s.loading {
		return "verifying…"
	}
	if s.err != nil {
		return "esc to cancel"
	}
	return ""
}

func (s *authStep) CanAdvance() (bool, string) {
	if s.loading {
		return false, "still verifying credentials"
	}
	if s.err != nil {
		return false, "fix AWS credentials first"
	}
	return true, ""
}

// ─── Inputs & Secrets ───────────────────────────────────────────────────────

type inputField struct {
	key      string
	desc     string
	secret   bool
	required bool
	target   *string // points into cfg.InstallInputs or cfg.Secrets[k].Value via a wrapper
	input    textinput.Model
}

type inputsStep struct {
	cfg    *stack.Config
	fields []*inputField
	cursor int

	// railOffset is the first rail line currently visible. Auto-adjusted in
	// renderRail to keep the focused field on screen; lets the rail scroll
	// when the field list overflows the body height.
	railOffset int

	// secret values live in the SecretInput struct (value type). We hold a
	// parallel map of name -> current edit value so we can write back into
	// cfg.Secrets on apply.
	secretEdits map[string]string
	inputEdits  map[string]string
}

func newInputsStep(cfg *stack.Config) *inputsStep {
	s := &inputsStep{
		cfg:         cfg,
		secretEdits: map[string]string{},
		inputEdits:  map[string]string{},
	}

	for _, k := range sortedKeys(cfg.InstallInputs) {
		v := cfg.InstallInputs[k]
		s.inputEdits[k] = v
		ti := textinput.New()
		ti.SetValue(v)
		ti.Placeholder = "(empty)"
		s.fields = append(s.fields, &inputField{
			key:    k,
			target: refInput(s.inputEdits, k),
			input:  ti,
		})
	}
	for _, k := range sortedKeys(cfg.Secrets) {
		sec := cfg.Secrets[k]
		s.secretEdits[k] = sec.Value
		ti := textinput.New()
		ti.SetValue(sec.Value)
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
		ti.Placeholder = "(empty)"
		s.fields = append(s.fields, &inputField{
			key:      k,
			desc:     sec.Description,
			secret:   true,
			required: sec.Required,
			target:   refInput(s.secretEdits, k),
			input:    ti,
		})
	}
	if len(s.fields) > 0 {
		_ = s.fields[0].input.Focus()
	}
	return s
}

func refInput(m map[string]string, k string) *string {
	// We don't take pointers into the map (Go forbids it). Targets are read
	// back on apply via the maps, not via this pointer. Keep nil here; the
	// apply path reads directly from the maps.
	_ = m
	_ = k
	return nil
}

func (s *inputsStep) Init() tea.Cmd { return textinput.Blink }

// focusables: 0 … len(fields)-1 = fields, then Previous, then Next.
func (s *inputsStep) prevIdx() int { return len(s.fields) }
func (s *inputsStep) nextIdx() int { return len(s.fields) + 1 }
func (s *inputsStep) onField() bool {
	return s.cursor < len(s.fields)
}

func (s *inputsStep) setCursor(i int) {
	max := len(s.fields) + 1 // Next
	if i < 0 {
		i = 0
	}
	if i > max {
		i = max
	}
	if s.onField() {
		s.fields[s.cursor].input.Blur()
	}
	s.cursor = i
	if s.onField() {
		_ = s.fields[s.cursor].input.Focus()
	}
}

func (s *inputsStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "down":
			s.setCursor(s.cursor + 1)
			return s, textinput.Blink
		case "up":
			s.setCursor(s.cursor - 1)
			return s, textinput.Blink
		case "left":
			// On a button, ←/→ toggle between Previous and Next so the user
			// can swap target without going back through every field. While
			// on a field, ← still belongs to the textinput cursor.
			if !s.onField() {
				s.setCursor(s.prevIdx())
				return s, nil
			}
		case "right":
			if !s.onField() {
				s.setCursor(s.nextIdx())
				return s, nil
			}
		case "enter":
			switch s.cursor {
			case s.prevIdx():
				return s, func() tea.Msg { return prevStepMsg{} }
			case s.nextIdx():
				return s, func() tea.Msg { return nextStepMsg{} }
			}
			return s, nil
		}
	}
	if !s.onField() {
		return s, nil
	}
	var cmd tea.Cmd
	s.fields[s.cursor].input, cmd = s.fields[s.cursor].input.Update(msg)
	f := s.fields[s.cursor]
	if f.secret {
		s.secretEdits[f.key] = f.input.Value()
	} else {
		s.inputEdits[f.key] = f.input.Value()
	}
	return s, cmd
}

func (s *inputsStep) applyEdits() {
	for k, v := range s.inputEdits {
		s.cfg.InstallInputs[k] = v
	}
	for k, v := range s.secretEdits {
		sec := s.cfg.Secrets[k]
		sec.Value = v
		s.cfg.Secrets[k] = sec
	}
}

func (s *inputsStep) View(w, h int) string {
	title := titleStyle.Render("Inputs & Secrets")
	if len(s.fields) == 0 {
		return title + "\n\n" +
			dimStyle.Render("No inputs or secrets to configure for this install.") + "\n\n" +
			s.renderButtons()
	}
	form := s.renderForm(w, h-4) // -4 for title + spacing + button row
	return title + "  " + s.renderProgress() + "\n\n" + form + "\n" + s.renderButtons()
}

func (s *inputsStep) renderButtons() string {
	prev := renderButton(" ◂ Previous ", s.cursor == s.prevIdx(), false)
	next := renderButton(" Next ▸ ", s.cursor == s.nextIdx(), false)
	return lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", next)
}

func (s *inputsStep) renderProgress() string {
	required, filled := 0, 0
	for _, f := range s.fields {
		if !f.required {
			continue
		}
		required++
		if strings.TrimSpace(s.secretEdits[f.key]) != "" {
			filled++
		}
	}
	if required == 0 {
		return dimStyle.Render(fmt.Sprintf("(%d optional fields)", len(s.fields)))
	}
	return dimStyle.Render(fmt.Sprintf("(%d/%d required filled)", filled, required))
}

// renderForm draws all fields top-to-bottom as a single vertical form. The
// focused field is highlighted; its description (if any) is shown dimmed
// directly beneath its input. Long forms scroll via s.railOffset so the
// focused field stays in view.
func (s *inputsStep) renderForm(width, height int) string {
	inputW := width - 6
	if inputW < 20 {
		inputW = width
	}

	var lines []string
	fieldStart := make([]int, len(s.fields))
	wroteInputs, wroteSecrets := false, false
	for i, f := range s.fields {
		if !f.secret && !wroteInputs {
			lines = append(lines, dimStyle.Render("INPUTS"))
			wroteInputs = true
		}
		if f.secret && !wroteSecrets {
			if wroteInputs {
				lines = append(lines, "")
			}
			lines = append(lines, dimStyle.Render("SECRETS"))
			wroteSecrets = true
		}

		focused := i == s.cursor
		marker := "  "
		label := f.key
		if focused {
			marker = focusedStyle.Render("▸ ")
			label = focusedStyle.Render(label)
		}
		if f.required {
			label += " " + requiredStyle.Render("(required)")
		}

		fieldStart[i] = len(lines)
		lines = append(lines, marker+label)

		f.input.SetWidth(inputW)
		lines = append(lines, "    "+f.input.View())

		if focused && f.desc != "" {
			for _, dl := range strings.Split(wrap(f.desc, width-6), "\n") {
				lines = append(lines, "    "+dimStyle.Render(dl))
			}
		}
		if focused && f.required && strings.TrimSpace(f.input.Value()) == "" {
			lines = append(lines, "    "+requiredStyle.Render("value required before continuing"))
		}
		lines = append(lines, "")
	}

	visible := height
	if visible < 1 {
		visible = 1
	}
	// Auto-scroll so the focused field's first line stays in view.
	if s.onField() && len(fieldStart) > 0 {
		target := fieldStart[s.cursor]
		if target < s.railOffset {
			s.railOffset = target
		} else if target >= s.railOffset+visible {
			s.railOffset = target - visible + 1
		}
	}
	if s.railOffset > len(lines)-visible {
		s.railOffset = len(lines) - visible
	}
	if s.railOffset < 0 {
		s.railOffset = 0
	}

	end := s.railOffset + visible
	if end > len(lines) {
		end = len(lines)
	}
	window := lines[s.railOffset:end]
	if s.railOffset > 0 {
		window = append([]string{dimStyle.Render(fmt.Sprintf("↑ %d more", s.railOffset))}, window[1:]...)
	}
	if end < len(lines) {
		window = append(window[:len(window)-1], dimStyle.Render(fmt.Sprintf("↓ %d more", len(lines)-end)))
	}

	return lipgloss.NewStyle().Width(width).Height(height).Render(strings.Join(window, "\n"))
}

func (s *inputsStep) Help() string {
	return "↑↓ move · type to edit"
}

func (s *inputsStep) CanAdvance() (bool, string) {
	for _, f := range s.fields {
		if f.required && strings.TrimSpace(s.secretEdits[f.key]) == "" {
			return false, fmt.Sprintf("required secret %q is empty", f.key)
		}
	}
	s.applyEdits()
	return true, ""
}

// ─── Network ────────────────────────────────────────────────────────────────

type networkStep struct {
	cfg    *stack.Config
	cursor int // 0 = Previous, 1 = Next
}

func newNetworkStep(cfg *stack.Config) *networkStep { return &networkStep{cfg: cfg, cursor: 1} }
func (s *networkStep) Init() tea.Cmd                { return nil }

func (s *networkStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	switch k.String() {
	case "left", "h", "up":
		s.cursor = 0
	case "right", "l", "down":
		s.cursor = 1
	case "enter":
		if s.cursor == 0 {
			return s, func() tea.Msg { return prevStepMsg{} }
		}
		return s, func() tea.Msg { return nextStepMsg{} }
	}
	return s, nil
}

func (s *networkStep) View(w, h int) string {
	title := titleStyle.Render("Network")
	diagram := networkDiagram()
	side := strings.Join([]string{
		kvRow("VPC CIDR", "10.128.0.0/16"),
		kvRow("Public subnets", "10.128.0.0/24, 10.128.16.0/24"),
		kvRow("Private subnets", "10.128.1.0/24, 10.128.17.0/24"),
		kvRow("Runner subnet", "10.128.2.0/24 (private)"),
		kvRow("NAT", "single, first public AZ"),
		"",
		dimStyle.Render("Existing VPC tagged for this install"),
		dimStyle.Render("will be adopted; otherwise created."),
	}, "\n")
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(diagram),
		"    ",
		side,
	)
	prev := renderButton(" ◂ Previous ", s.cursor == 0, false)
	next := renderButton(" Next ▸ ", s.cursor == 1, false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", next)
	return title + "\n\n" + body + "\n\n" + buttons
}

func (s *networkStep) Help() string               { return "←→ move" }
func (s *networkStep) CanAdvance() (bool, string) { return true, "" }

func networkDiagram() string {
	return `┌──────────── VPC 10.128.0.0/16 ────────────┐
│                                            │
│   AZ-a                  AZ-b               │
│  ┌──────────┐         ┌──────────┐         │
│  │ public   │         │ public   │  ◀ IGW │
│  │ .0.0/24  │         │ .16.0/24 │         │
│  └────┬─────┘         └──────────┘         │
│       │ NAT                                │
│  ┌────▼─────┐         ┌──────────┐         │
│  │ private  │         │ private  │         │
│  │ .1.0/24  │         │ .17.0/24 │         │
│  └──────────┘         └──────────┘         │
│  ┌──────────┐                              │
│  │ runner   │ .2.0/24 (private)            │
│  └──────────┘                              │
└────────────────────────────────────────────┘`
}

// ─── Roles ──────────────────────────────────────────────────────────────────

type roleGroupKind int

const (
	grpOps roleGroupKind = iota
	grpBreakGlass
	grpCustom
)

type roleEntry struct {
	key     string // map key or op role identifier
	label   string // display label
	enabled bool
	// For ops roles: snapshot of original policy fields to restore on re-enable.
	opSnapshot *opSnapshot
}

type opSnapshot struct {
	inline     string
	perms      []string
	managedArn []string
}

type rolesStep struct {
	cfg    *stack.Config
	groups [3][]*roleEntry // grpOps, grpBreakGlass, grpCustom
	// cursor is a flat index over [all entries across non-empty groups...] then
	// Previous button, then Next button. Walked linearly by up/down so the user
	// moves through panels in reading order before reaching the buttons.
	cursor int
}

// totalEntries returns the total number of role entries across all groups.
func (s *rolesStep) totalEntries() int {
	return len(s.groups[grpOps]) + len(s.groups[grpBreakGlass]) + len(s.groups[grpCustom])
}

// resolveCursor maps the flat cursor to (group, indexInGroup) for entry slots,
// or returns isButton=true with btn ∈ {"prev", "next"} for the two trailing slots.
func (s *rolesStep) resolveCursor() (grp roleGroupKind, idx int, isButton bool, btn string) {
	i := s.cursor
	for g := range s.groups {
		if i < len(s.groups[g]) {
			return roleGroupKind(g), i, false, ""
		}
		i -= len(s.groups[g])
	}
	if i == 0 {
		return 0, 0, true, "prev"
	}
	return 0, 0, true, "next"
}

func (s *rolesStep) prevCursorIdx() int { return s.totalEntries() }
func (s *rolesStep) nextCursorIdx() int { return s.totalEntries() + 1 }

func newRolesStep(cfg *stack.Config) *rolesStep {
	s := &rolesStep{cfg: cfg}

	ops := []struct {
		key   string
		label string
		on    bool
		snap  *opSnapshot
	}{
		{"provision", "Provision (install)", hasOp(cfg.ProvisionInlinePolicyDocument, cfg.ProvisionPermissions, cfg.ProvisionManagedPolicyARNs),
			&opSnapshot{cfg.ProvisionInlinePolicyDocument, cfg.ProvisionPermissions, cfg.ProvisionManagedPolicyARNs}},
		{"maintenance", "Maintenance (day-2 ops)", hasOp(cfg.MaintenanceInlinePolicyDocument, cfg.MaintenancePermissions, cfg.MaintenanceManagedPolicyARNs),
			&opSnapshot{cfg.MaintenanceInlinePolicyDocument, cfg.MaintenancePermissions, cfg.MaintenanceManagedPolicyARNs}},
		{"deprovision", "Deprovision (teardown)", hasOp(cfg.DeprovisionInlinePolicyDocument, cfg.DeprovisionPermissions, cfg.DeprovisionManagedPolicyARNs),
			&opSnapshot{cfg.DeprovisionInlinePolicyDocument, cfg.DeprovisionPermissions, cfg.DeprovisionManagedPolicyARNs}},
	}
	for _, o := range ops {
		s.groups[grpOps] = append(s.groups[grpOps], &roleEntry{
			key: o.key, label: o.label, enabled: o.on, opSnapshot: o.snap,
		})
	}
	for _, k := range sortedKeys(cfg.BreakGlassRoles) {
		v := cfg.BreakGlassRoles[k]
		s.groups[grpBreakGlass] = append(s.groups[grpBreakGlass], &roleEntry{
			key: k, label: k, enabled: v.Enabled,
		})
	}
	for _, k := range sortedKeys(cfg.CustomRoles) {
		v := cfg.CustomRoles[k]
		s.groups[grpCustom] = append(s.groups[grpCustom], &roleEntry{
			key: k, label: k, enabled: v.Enabled,
		})
	}
	return s
}

func hasOp(inline string, perms, arns []string) bool {
	return inline != "" || len(perms) > 0 || len(arns) > 0
}

func (s *rolesStep) Init() tea.Cmd { return nil }

func (s *rolesStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	max := s.nextCursorIdx()
	switch k.String() {
	case "down", "j":
		if s.cursor < max {
			s.cursor++
		}
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "right", "l":
		// Move to Next button quickly; mirrors left/right between buttons.
		if s.cursor == s.prevCursorIdx() {
			s.cursor = s.nextCursorIdx()
		}
	case "left", "h":
		if s.cursor == s.nextCursorIdx() {
			s.cursor = s.prevCursorIdx()
		}
	case " ", "x":
		_, idx, isBtn, _ := s.resolveCursor()
		if isBtn {
			return s, nil
		}
		grp, _, _, _ := s.resolveCursor()
		s.groups[grp][idx].enabled = !s.groups[grp][idx].enabled
	case "enter":
		_, _, isBtn, btn := s.resolveCursor()
		if !isBtn {
			return s, nil
		}
		s.apply()
		if btn == "prev" {
			return s, func() tea.Msg { return prevStepMsg{} }
		}
		return s, func() tea.Msg { return nextStepMsg{} }
	}
	return s, nil
}

func (s *rolesStep) apply() {
	// Operation roles: zero or restore.
	for _, r := range s.groups[grpOps] {
		switch r.key {
		case "provision":
			if r.enabled {
				s.cfg.ProvisionInlinePolicyDocument = r.opSnapshot.inline
				s.cfg.ProvisionPermissions = r.opSnapshot.perms
				s.cfg.ProvisionManagedPolicyARNs = r.opSnapshot.managedArn
			} else {
				s.cfg.ProvisionInlinePolicyDocument = ""
				s.cfg.ProvisionPermissions = nil
				s.cfg.ProvisionManagedPolicyARNs = nil
			}
		case "maintenance":
			if r.enabled {
				s.cfg.MaintenanceInlinePolicyDocument = r.opSnapshot.inline
				s.cfg.MaintenancePermissions = r.opSnapshot.perms
				s.cfg.MaintenanceManagedPolicyARNs = r.opSnapshot.managedArn
			} else {
				s.cfg.MaintenanceInlinePolicyDocument = ""
				s.cfg.MaintenancePermissions = nil
				s.cfg.MaintenanceManagedPolicyARNs = nil
			}
		case "deprovision":
			if r.enabled {
				s.cfg.DeprovisionInlinePolicyDocument = r.opSnapshot.inline
				s.cfg.DeprovisionPermissions = r.opSnapshot.perms
				s.cfg.DeprovisionManagedPolicyARNs = r.opSnapshot.managedArn
			} else {
				s.cfg.DeprovisionInlinePolicyDocument = ""
				s.cfg.DeprovisionPermissions = nil
				s.cfg.DeprovisionManagedPolicyARNs = nil
			}
		}
	}
	for _, r := range s.groups[grpBreakGlass] {
		v := s.cfg.BreakGlassRoles[r.key]
		v.Enabled = r.enabled
		s.cfg.BreakGlassRoles[r.key] = v
	}
	for _, r := range s.groups[grpCustom] {
		v := s.cfg.CustomRoles[r.key]
		v.Enabled = r.enabled
		s.cfg.CustomRoles[r.key] = v
	}
}

func (s *rolesStep) View(w, h int) string {
	title := titleStyle.Render("Roles")

	panels := []string{
		s.renderPanel("Operation roles", "Disable to skip creating; policy template is remembered.", grpOps),
		s.renderPanel("Break-glass roles", "Off by default. Enable for incident access.", grpBreakGlass),
		s.renderPanel("Custom roles", "App-defined roles. Toggle per stack.", grpCustom),
	}

	prev := renderButton(" ◂ Previous ", s.cursor == s.prevCursorIdx(), false)
	next := renderButton(" Next ▸ ", s.cursor == s.nextCursorIdx(), false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", next)

	// Buttons live in the left column, directly under the role panels, so a
	// tall detail pane on the right doesn't push them to the bottom.
	left := strings.Join(panels, "\n") + "\n" + buttons

	// Detail pane fills the remaining width after the left column + a 4-col
	// gutter. Fall back to w/3 if the math comes out too narrow.
	leftW := lipgloss.Width(left)
	detailW := w - leftW - 4
	if detailW < w/3 {
		detailW = w / 3
	}
	detail := s.renderDetail(detailW)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, "    ", detail)
	return title + "\n\n" + body
}

func (s *rolesStep) renderPanel(title, desc string, k roleGroupKind) string {
	on := 0
	for _, r := range s.groups[k] {
		if r.enabled {
			on++
		}
	}
	curGrp, curIdx, isBtn, _ := s.resolveCursor()
	groupHasFocus := !isBtn && curGrp == k
	header := "  " + kvKeyStyle.Render(title) + "  " + dimStyle.Render(fmt.Sprintf("(%d/%d enabled)", on, len(s.groups[k])))
	lines := []string{header, dimStyle.Render("  " + desc)}
	if len(s.groups[k]) == 0 {
		lines = append(lines, dimStyle.Render("    (none defined)"))
	}
	for i, r := range s.groups[k] {
		box := checkboxOff.Render("[ ]")
		if r.enabled {
			box = checkboxOn.Render("[x]")
		}
		focused := groupHasFocus && i == curIdx
		marker := "  "
		if focused {
			marker = focusedStyle.Render("▸ ")
		}
		label := r.label
		if focused {
			label = focusedStyle.Render(r.label)
		}
		row := fmt.Sprintf("  %s%s %s", marker, box, label)
		lines = append(lines, row)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (s *rolesStep) renderDetail(width int) string {
	if width < 20 {
		return ""
	}
	curGrp, curIdx, isBtn, _ := s.resolveCursor()
	if isBtn {
		return dimStyle.Render("Move ↑ back into a role to see its details.")
	}
	entries := s.groups[curGrp]
	if len(entries) == 0 {
		return dimStyle.Render("No roles in this group.")
	}
	r := entries[curIdx]
	lines := []string{kvKeyStyle.Render("Details"), ""}
	switch curGrp {
	case grpOps:
		lines = append(lines, kvRow("Role", r.label))
		lines = append(lines, kvRow("State", onOff(r.enabled)))
		if r.opSnapshot != nil {
			lines = append(lines, kvRow("Permissions", fmt.Sprintf("%d", len(r.opSnapshot.perms))))
			lines = append(lines, kvRow("Managed ARNs", fmt.Sprintf("%d", len(r.opSnapshot.managedArn))))
			if r.opSnapshot.inline != "" {
				lines = append(lines, "", dimStyle.Render("Inline policy (preview):"))
				lines = append(lines, prettyJSON(r.opSnapshot.inline, width-2, 16))
			}
		}
	case grpBreakGlass:
		v := s.cfg.BreakGlassRoles[r.key]
		lines = append(lines, kvRow("Role", r.key))
		lines = append(lines, kvRow("State", onOff(r.enabled)))
		lines = append(lines, kvRow("Permissions", fmt.Sprintf("%d", len(v.Permissions))))
		lines = append(lines, kvRow("Managed ARNs", fmt.Sprintf("%d", len(v.ManagedPolicyARNs))))
		if v.InlinePolicyDocument != "" {
			lines = append(lines, "", dimStyle.Render("Inline policy (preview):"))
			lines = append(lines, prettyJSON(v.InlinePolicyDocument, width-2, 16))
		}
	case grpCustom:
		v := s.cfg.CustomRoles[r.key]
		lines = append(lines, kvRow("Role", r.key))
		lines = append(lines, kvRow("State", onOff(r.enabled)))
		lines = append(lines, kvRow("Permissions", fmt.Sprintf("%d", len(v.Permissions))))
		lines = append(lines, kvRow("Managed ARNs", fmt.Sprintf("%d", len(v.ManagedPolicyARNs))))
		if v.InlinePolicyDocument != "" {
			lines = append(lines, "", dimStyle.Render("Inline policy (preview):"))
			lines = append(lines, prettyJSON(v.InlinePolicyDocument, width-2, 16))
		}
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n"))
}

func (s *rolesStep) Help() string {
	return "↑↓ move · space toggle"
}

func (s *rolesStep) CanAdvance() (bool, string) {
	s.apply()
	return true, ""
}

// ─── Confirm ────────────────────────────────────────────────────────────────

type confirmStep struct {
	cfg       *stack.Config
	accountID string // populated by the app when entering this step
	cursor    int    // 0 = Previous, 1 = Provision
}

func newConfirmStep(cfg *stack.Config) *confirmStep { return &confirmStep{cfg: cfg, cursor: 1} }
func (s *confirmStep) Init() tea.Cmd                { return nil }

func (s *confirmStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	switch k.String() {
	case "left", "h", "up":
		s.cursor = 0
	case "right", "l", "down":
		s.cursor = 1
	case "enter":
		if s.cursor == 0 {
			return s, func() tea.Msg { return prevStepMsg{} }
		}
		return s, func() tea.Msg { return finishMsg{} }
	}
	return s, nil
}

func (s *confirmStep) View(w, h int) string {
	title := titleStyle.Render("Provision")

	opRoles := []string{}
	if s.cfg.ProvisionInlinePolicyDocument != "" || len(s.cfg.ProvisionPermissions) > 0 || len(s.cfg.ProvisionManagedPolicyARNs) > 0 {
		opRoles = append(opRoles, "provision")
	}
	if s.cfg.MaintenanceInlinePolicyDocument != "" || len(s.cfg.MaintenancePermissions) > 0 || len(s.cfg.MaintenanceManagedPolicyARNs) > 0 {
		opRoles = append(opRoles, "maintenance")
	}
	if s.cfg.DeprovisionInlinePolicyDocument != "" || len(s.cfg.DeprovisionPermissions) > 0 || len(s.cfg.DeprovisionManagedPolicyARNs) > 0 {
		opRoles = append(opRoles, "deprovision")
	}
	bgOn := 0
	for _, v := range s.cfg.BreakGlassRoles {
		if v.Enabled {
			bgOn++
		}
	}
	crOn := 0
	for _, v := range s.cfg.CustomRoles {
		if v.Enabled {
			crOn++
		}
	}

	var leftRows []string
	if s.accountID != "" {
		leftRows = append(leftRows, kvRow("Account", s.accountID))
	}
	leftRows = append(leftRows, kvRow("Region", s.cfg.AWSRegion))
	leftRows = append(leftRows, kvRow("Method", methodLabel(s.cfg.Method)))
	left := strings.Join(leftRows, "\n")
	right := strings.Join([]string{
		kvRow("Inputs", fmt.Sprintf("%d", len(s.cfg.InstallInputs))),
		kvRow("Secrets", fmt.Sprintf("%d (%d auto)", len(s.cfg.Secrets), len(s.cfg.AutoGenerateSecrets))),
		kvRow("Op roles", strings.Join(opRoles, ", ")),
		kvRow("Break-glass", fmt.Sprintf("%d / %d", bgOn, len(s.cfg.BreakGlassRoles))),
		kvRow("Custom", fmt.Sprintf("%d / %d", crOn, len(s.cfg.CustomRoles))),
	}, "\n")

	cols := lipgloss.JoinHorizontal(lipgloss.Top, left, "        ", right)
	warn := dimStyle.Render("AWS resources will be created in your account. Activate Provision to begin.")

	prev := renderButton(" ◂ Previous ", s.cursor == 0, false)
	provision := renderButton(" Provision ✓ ", s.cursor == 1, false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", provision)
	return title + "\n\n" + cols + "\n\n" + warn + "\n\n" + buttons
}

func (s *confirmStep) Help() string               { return "←→ move" }
func (s *confirmStep) CanAdvance() (bool, string) { return true, "" }

// ─── helpers ────────────────────────────────────────────────────────────────

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func kvRow(k, v string) string {
	if v == "" {
		v = dimStyle.Render("—")
	}
	return kvKeyStyle.Render(fmt.Sprintf("%-14s", k+":")) + "  " + v
}

func methodLabel(m stack.Method) string {
	switch m {
	case stack.MethodTerraform:
		return "Terraform module"
	case stack.MethodSDK:
		return "AWS SDK"
	default:
		return "AWS SDK (default)"
	}
}

func onOff(b bool) string {
	if b {
		return checkboxOn.Render("enabled")
	}
	return checkboxOff.Render("disabled")
}

func truncate(s string, max int) string {
	if max <= 1 || len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func wrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var out []string
	for _, line := range strings.Split(s, "\n") {
		for len(line) > width {
			cut := width
			if idx := strings.LastIndex(line[:width], " "); idx > 0 {
				cut = idx
			}
			out = append(out, line[:cut])
			line = strings.TrimLeft(line[cut:], " ")
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func truncBlock(s string, width, lines int) string {
	wrapped := wrap(s, width)
	parts := strings.Split(wrapped, "\n")
	if len(parts) > lines {
		parts = parts[:lines]
		parts = append(parts, dimStyle.Render("…"))
	}
	return strings.Join(parts, "\n")
}

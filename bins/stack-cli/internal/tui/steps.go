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
	kind  stack.Kind
	cloud stack.Cloud
}

func newIntroStep(kind stack.Kind, cloud stack.Cloud) *introStep {
	return &introStep{kind: kind, cloud: cloud}
}
func (s *introStep) Init() tea.Cmd { return nil }

func (s *introStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "enter" {
		return s, func() tea.Msg { return nextStepMsg{} }
	}
	return s, nil
}

func (s *introStep) Main(w, h int) string {
	verb := "provision"
	switch s.kind {
	case stack.KindReprovision:
		verb = "reprovision"
	case stack.KindDeprovision:
		verb = "deprovision"
	}

	if s.cloud == stack.CloudGCP {
		return s.mainGCP(verb)
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
			"  " + stepActive.Render("1.") + " " + kvKeyStyle.Render("Auth") + "     — verify your AWS credentials and region",
			"  " + stepActive.Render("2.") + " " + kvKeyStyle.Render("Inputs") + "   — fill in install inputs",
			"  " + stepActive.Render("3.") + " " + kvKeyStyle.Render("Secrets") + "  — set required secrets",
			"  " + stepActive.Render("4.") + " " + kvKeyStyle.Render("Network") + "  — review the VPC plan",
			"  " + stepActive.Render("5.") + " " + kvKeyStyle.Render("Roles") + "    — pick which IAM roles to create",
			"  " + stepActive.Render("6.") + " " + kvKeyStyle.Render("Provision") + " — confirm the plan and apply it",
		}, "\n")
	}

	next := renderButton(" Get started ▸ ", true, false)
	return heading + "\n\n" + body + "\n\n" + next
}

func (s *introStep) mainGCP(verb string) string {
	heading := titleStyle.Render("Welcome")
	var body string
	switch s.kind {
	case stack.KindDeprovision:
		body = strings.Join([]string{
			"A " + kvKeyStyle.Render("stack") + " is the set of GCP resources installed in your",
			"project: a VPC network and subnets, service accounts and",
			"IAM bindings, Secret Manager entries, and a runner GCE",
			"instance. Applications run inside this stack.",
			"",
			kvKeyStyle.Render("Deprovision") + " tears everything down via Terraform, leaving",
			"the project empty of stack-managed resources.",
		}, "\n")
	default:
		body = strings.Join([]string{
			"A " + kvKeyStyle.Render("stack") + " is the set of GCP resources installed in your",
			"project so applications can run there: a VPC network and",
			"subnets, service accounts and IAM bindings, Secret Manager",
			"entries, and a runner GCE instance.",
			"",
			"This wizard walks you through " + verb + "ing one. You'll:",
			"",
			"  " + stepActive.Render("1.") + " " + kvKeyStyle.Render("Inputs") + "   — fill in install inputs",
			"  " + stepActive.Render("2.") + " " + kvKeyStyle.Render("Secrets") + "  — set required secrets",
			"  " + stepActive.Render("3.") + " " + kvKeyStyle.Render("Roles") + "    — pick which IAM roles to create",
			"  " + stepActive.Render("4.") + " " + kvKeyStyle.Render("Auth") + "     — verify your GCP credentials",
			"  " + stepActive.Render("5.") + " " + kvKeyStyle.Render("Network") + "  — set the project/region and review the VPC plan",
			"  " + stepActive.Render("6.") + " " + kvKeyStyle.Render("Provision") + " — confirm the plan and apply it",
		}, "\n")
	}
	next := renderButton(" Get started ▸ ", true, false)
	return heading + "\n\n" + body + "\n\n" + next
}

func (s *introStep) Detail(w, h int) string {
	if s.cloud == stack.CloudGCP {
		if s.kind == stack.KindDeprovision {
			return strings.Join([]string{
				kvKeyStyle.Render("This will delete"),
				"",
				dimStyle.Render("· the runner GCE instance and logs"),
				dimStyle.Render("· all created service accounts and IAM bindings"),
				dimStyle.Render("· stack secrets in GCP Secret Manager"),
				dimStyle.Render("· the VPC network, subnets, and router/NAT"),
			}, "\n")
		}
		return strings.Join([]string{
			kvKeyStyle.Render("Before you begin"),
			"",
			dimStyle.Render(wrap("Nothing is created in your GCP project until you reach the final step and select Provision.", w)),
		}, "\n")
	}
	if s.kind == stack.KindDeprovision {
		return strings.Join([]string{
			kvKeyStyle.Render("This will delete"),
			"",
			dimStyle.Render("· the runner EC2 instance and log group"),
			dimStyle.Render("· all created IAM roles and policies"),
			dimStyle.Render("· stack secrets in AWS Secrets Manager"),
			dimStyle.Render("· the VPC, subnets, NAT gateway, and route tables"),
		}, "\n")
	}
	return strings.Join([]string{
		kvKeyStyle.Render("Before you begin"),
		"",
		dimStyle.Render(wrap("Nothing is created in your AWS account until you reach the final step and select Provision.", w)),
	}, "\n")
}

func (s *introStep) Help() string               { return "" }
func (s *introStep) CanAdvance() (bool, string) { return true, "" }

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
	awsCfg, err := awsconfig.LoadDefaultConfig(s.ctx, awsconfig.WithRegion(s.cfg.AWS.Region))
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

func (s *authStep) Main(w, h int) string {
	title := titleStyle.Render("Verify AWS credentials")
	var body string
	switch {
	case s.loading:
		body = s.spinner.View() + " " + dimStyle.Render("Calling sts:GetCallerIdentity…")
	case s.err != nil:
		body = errStyle.Render("Error: " + s.err.Error())
	default:
		rows := []string{
			kvRow("Account", s.account),
			kvRow("ARN", s.arn),
			kvRow("Region", s.cfg.AWS.Region),
		}
		body = strings.Join(rows, "\n") + "\n\n" + renderButton(" Continue ▸ ", true, false)
	}
	return title + "\n\n" + body
}

func (s *authStep) Detail(w, h int) string {
	switch {
	case s.loading:
		return dimStyle.Render(wrap("Confirming which AWS account and principal this stack will be provisioned into.", w))
	case s.err != nil:
		return kvKeyStyle.Render("Troubleshooting") + "\n\n" +
			dimStyle.Render(wrap("Fix your AWS credentials (env, profile, or SSO) and restart. If using SSO, run aws sso login first.", w))
	default:
		return kvKeyStyle.Render("Identity") + "\n\n" +
			dimStyle.Render(wrap("These credentials will be used for every AWS call during provisioning.", w))
	}
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
	cfg      *stack.Config
	fields   []*inputField
	cursor   int
	title    string
	emptyMsg string

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

// newInputsStep builds the install-inputs step (no secrets).
func newInputsStep(cfg *stack.Config) *inputsStep {
	return newFieldStep(cfg, true, false, "Inputs", "No inputs to configure for this install.")
}

// newSecretsStep builds the secrets step (no install inputs).
func newSecretsStep(cfg *stack.Config) *inputsStep {
	return newFieldStep(cfg, false, true, "Secrets", "No secrets to configure for this install.")
}

// newFieldStep is the shared builder behind the inputs and secrets steps; both
// are the same textinput-driven form over a filtered field set.
func newFieldStep(cfg *stack.Config, includeInputs, includeSecrets bool, title, emptyMsg string) *inputsStep {
	s := &inputsStep{
		cfg:         cfg,
		secretEdits: map[string]string{},
		inputEdits:  map[string]string{},
		title:       title,
		emptyMsg:    emptyMsg,
	}

	if includeInputs {
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
	}
	if includeSecrets {
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

func (s *inputsStep) Main(w, h int) string {
	title := titleStyle.Render(s.title)
	if len(s.fields) == 0 {
		return title + "\n\n" +
			dimStyle.Render(s.emptyMsg) + "\n\n" +
			s.renderButtons()
	}
	form := s.renderForm(w, h-4) // -4 for title + spacing + button row
	return title + "  " + s.renderProgress() + "\n\n" + form + "\n" + s.renderButtons()
}

func (s *inputsStep) Detail(w, h int) string {
	if len(s.fields) == 0 {
		return ""
	}
	if !s.onField() {
		return dimStyle.Render(wrap("Move ↑ into a field to see its details.", w))
	}
	f := s.fields[s.cursor]
	kind := "input"
	if f.secret {
		kind = "secret"
	}
	lines := []string{
		kvKeyStyle.Render("Details"), "",
		kvRow("Field", f.key),
		kvRow("Type", kind),
		kvRow("Required", yesNo(f.required)),
	}
	if f.secret {
		lines = append(lines, kvRow("Length", fmt.Sprintf("%d chars", len(f.input.Value()))))
	} else {
		lines = append(lines, kvRow("Value", f.input.Value()))
	}
	if f.desc != "" {
		lines = append(lines, "", dimStyle.Render(wrap(f.desc, w)))
	}
	return strings.Join(lines, "\n")
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
	for i, f := range s.fields {
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

		if focused && f.required && strings.TrimSpace(f.input.Value()) == "" {
			lines = append(lines, "    "+requiredStyle.Render("required — must be set before provisioning"))
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
	// Required fields are not enforced here — the user can move forward with
	// them empty and fill them later. Validation happens at provision time.
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

func (s *networkStep) Main(w, h int) string {
	title := titleStyle.Render("Network")
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
	prev := renderButton(" ◂ Previous ", s.cursor == 0, false)
	next := renderButton(" Next ▸ ", s.cursor == 1, false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", next)
	return title + "\n\n" + side + "\n\n" + buttons
}

func (s *networkStep) Detail(w, h int) string {
	return networkDiagram()
}

func (s *networkStep) Help() string               { return "←→ move" }
func (s *networkStep) CanAdvance() (bool, string) { return true, "" }

// networkDiagram composes the VPC topology from lipgloss-bordered boxes rather
// than hand-drawn ASCII so every border edge lines up regardless of label
// width. Boxes are joined vertically per-AZ and horizontally across AZs; the
// outer box carries the VPC title.
func networkDiagram() string {
	cyan := lipgloss.Color("39")
	lbl := lipgloss.NewStyle().Foreground(cyan).Bold(true)
	note := lipgloss.NewStyle().Foreground(cyan)

	box := func(name, cidr string) string {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cyan).
			Foreground(cyan).
			Padding(1, 2).
			Width(24).
			Render(fmt.Sprintf("%-8s %s", name, cidr))
	}

	// boxW is the rendered box width: inner 24 + padding 4 + border 2 = 30.
	const boxW = 30
	center := strings.Repeat(" ", boxW/2)
	nat := note.Render(center + "│\n" + strings.Repeat(" ", boxW/2-1) + "NAT\n" + center + "▼")

	colA := lipgloss.JoinVertical(lipgloss.Left,
		lbl.Render("AZ-a"),
		box("public", ".0.0/24"),
		nat,
		box("private", ".1.0/24"),
	)
	colB := lipgloss.JoinVertical(lipgloss.Left,
		lbl.Render("AZ-b"),
		box("public", ".16.0/24"),
		"", "", "", // pad to align private box with AZ-a's, past the NAT connector
		box("private", ".17.0/24"),
	)
	igw := lipgloss.JoinVertical(lipgloss.Left, "", "", "", lbl.Render("◀ IGW"))

	azRow := lipgloss.JoinHorizontal(lipgloss.Top, colA, "        ", colB, "      ", igw)

	runnerRow := lipgloss.JoinHorizontal(lipgloss.Center,
		box("runner", ".2.0/24"),
		note.Render("  (private)"),
	)

	inner := lipgloss.JoinVertical(lipgloss.Left,
		lbl.Render("VPC 10.128.0.0/16"),
		"",
		azRow,
		"",
		runnerRow,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cyan).
		Padding(1, 2).
		Render(inner)
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
		{"provision", "Provision (install)", hasOp(cfg.AWS.ProvisionInlinePolicyDocument, cfg.AWS.ProvisionPermissions, cfg.AWS.ProvisionManagedPolicyARNs),
			&opSnapshot{cfg.AWS.ProvisionInlinePolicyDocument, cfg.AWS.ProvisionPermissions, cfg.AWS.ProvisionManagedPolicyARNs}},
		{"maintenance", "Maintenance (day-2 ops)", hasOp(cfg.AWS.MaintenanceInlinePolicyDocument, cfg.AWS.MaintenancePermissions, cfg.AWS.MaintenanceManagedPolicyARNs),
			&opSnapshot{cfg.AWS.MaintenanceInlinePolicyDocument, cfg.AWS.MaintenancePermissions, cfg.AWS.MaintenanceManagedPolicyARNs}},
		{"deprovision", "Deprovision (teardown)", hasOp(cfg.AWS.DeprovisionInlinePolicyDocument, cfg.AWS.DeprovisionPermissions, cfg.AWS.DeprovisionManagedPolicyARNs),
			&opSnapshot{cfg.AWS.DeprovisionInlinePolicyDocument, cfg.AWS.DeprovisionPermissions, cfg.AWS.DeprovisionManagedPolicyARNs}},
	}
	for _, o := range ops {
		s.groups[grpOps] = append(s.groups[grpOps], &roleEntry{
			key: o.key, label: o.label, enabled: o.on, opSnapshot: o.snap,
		})
	}
	for _, k := range sortedKeys(cfg.AWS.BreakGlassRoles) {
		v := cfg.AWS.BreakGlassRoles[k]
		s.groups[grpBreakGlass] = append(s.groups[grpBreakGlass], &roleEntry{
			key: k, label: k, enabled: v.Enabled,
		})
	}
	for _, k := range sortedKeys(cfg.AWS.CustomRoles) {
		v := cfg.AWS.CustomRoles[k]
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
	case "space", "x":
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
				s.cfg.AWS.ProvisionInlinePolicyDocument = r.opSnapshot.inline
				s.cfg.AWS.ProvisionPermissions = r.opSnapshot.perms
				s.cfg.AWS.ProvisionManagedPolicyARNs = r.opSnapshot.managedArn
			} else {
				s.cfg.AWS.ProvisionInlinePolicyDocument = ""
				s.cfg.AWS.ProvisionPermissions = nil
				s.cfg.AWS.ProvisionManagedPolicyARNs = nil
			}
		case "maintenance":
			if r.enabled {
				s.cfg.AWS.MaintenanceInlinePolicyDocument = r.opSnapshot.inline
				s.cfg.AWS.MaintenancePermissions = r.opSnapshot.perms
				s.cfg.AWS.MaintenanceManagedPolicyARNs = r.opSnapshot.managedArn
			} else {
				s.cfg.AWS.MaintenanceInlinePolicyDocument = ""
				s.cfg.AWS.MaintenancePermissions = nil
				s.cfg.AWS.MaintenanceManagedPolicyARNs = nil
			}
		case "deprovision":
			if r.enabled {
				s.cfg.AWS.DeprovisionInlinePolicyDocument = r.opSnapshot.inline
				s.cfg.AWS.DeprovisionPermissions = r.opSnapshot.perms
				s.cfg.AWS.DeprovisionManagedPolicyARNs = r.opSnapshot.managedArn
			} else {
				s.cfg.AWS.DeprovisionInlinePolicyDocument = ""
				s.cfg.AWS.DeprovisionPermissions = nil
				s.cfg.AWS.DeprovisionManagedPolicyARNs = nil
			}
		}
	}
	for _, r := range s.groups[grpBreakGlass] {
		v := s.cfg.AWS.BreakGlassRoles[r.key]
		v.Enabled = r.enabled
		s.cfg.AWS.BreakGlassRoles[r.key] = v
	}
	for _, r := range s.groups[grpCustom] {
		v := s.cfg.AWS.CustomRoles[r.key]
		v.Enabled = r.enabled
		s.cfg.AWS.CustomRoles[r.key] = v
	}
}

func (s *rolesStep) Main(w, h int) string {
	title := titleStyle.Render("Roles")

	panels := []string{
		s.renderPanel("Operation roles", "Disable to skip creating; policy template is remembered.", grpOps),
		s.renderPanel("Break-glass roles", "Off by default. Enable for incident access.", grpBreakGlass),
		s.renderPanel("Custom roles", "App-defined roles. Toggle per stack.", grpCustom),
	}

	prev := renderButton(" ◂ Previous ", s.cursor == s.prevCursorIdx(), false)
	next := renderButton(" Next ▸ ", s.cursor == s.nextCursorIdx(), false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", next)

	return title + "\n\n" + strings.Join(panels, "\n") + "\n" + buttons
}

func (s *rolesStep) Detail(w, h int) string {
	return s.renderDetail(w)
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
		v := s.cfg.AWS.BreakGlassRoles[r.key]
		lines = append(lines, kvRow("Role", r.key))
		lines = append(lines, kvRow("State", onOff(r.enabled)))
		lines = append(lines, kvRow("Permissions", fmt.Sprintf("%d", len(v.Permissions))))
		lines = append(lines, kvRow("Managed ARNs", fmt.Sprintf("%d", len(v.ManagedPolicyARNs))))
		if v.InlinePolicyDocument != "" {
			lines = append(lines, "", dimStyle.Render("Inline policy (preview):"))
			lines = append(lines, prettyJSON(v.InlinePolicyDocument, width-2, 16))
		}
	case grpCustom:
		v := s.cfg.AWS.CustomRoles[r.key]
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

func (s *confirmStep) Main(w, h int) string {
	title := titleStyle.Render("Provision")

	var rows []string
	var warnText string
	if s.cfg.Cloud == stack.CloudGCP {
		g := s.cfg.GCP
		if g != nil {
			rows = append(rows, kvRow("Project", g.ProjectID))
			rows = append(rows, kvRow("Region", g.Region))
		}
		warnText = "GCP resources will be created in your project. Activate Provision to begin."
	} else {
		if s.accountID != "" {
			rows = append(rows, kvRow("Account", s.accountID))
		}
		if s.cfg.AWS != nil {
			rows = append(rows, kvRow("Region", s.cfg.AWS.Region))
		}
		warnText = "AWS resources will be created in your account. Activate Provision to begin."
	}

	warn := dimStyle.Render(wrap(warnText, w))

	prev := renderButton(" ◂ Previous ", s.cursor == 0, false)
	provision := renderButton(" Provision ✓ ", s.cursor == 1, false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", provision)
	return title + "\n\n" + strings.Join(rows, "\n") + "\n\n" + warn + "\n\n" + buttons
}

func (s *confirmStep) Detail(w, h int) string {
	var lines []string
	section := func(title string) {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, kvKeyStyle.Render(title))
	}
	item := func(label, value string) {
		lines = append(lines, "  "+dimStyle.Render(label)+"  "+value)
	}
	none := func() {
		lines = append(lines, "  "+dimStyle.Render("(none)"))
	}

	section(fmt.Sprintf("Inputs (%d)", len(s.cfg.InstallInputs)))
	if len(s.cfg.InstallInputs) == 0 {
		none()
	}
	for _, k := range sortedKeys(s.cfg.InstallInputs) {
		v := s.cfg.InstallInputs[k]
		if v == "" {
			v = dimStyle.Render("(empty)")
		} else {
			v = truncate(v, w/2)
		}
		item(k, v)
	}

	auto := map[string]bool{}
	for _, k := range s.cfg.AutoGenerateSecrets {
		auto[k] = true
	}
	section(fmt.Sprintf("Secrets (%d)", len(s.cfg.Secrets)))
	if len(s.cfg.Secrets) == 0 {
		none()
	}
	for _, k := range sortedKeys(s.cfg.Secrets) {
		sec := s.cfg.Secrets[k]
		var tag string
		switch {
		case auto[k]:
			tag = dimStyle.Render("auto-generated")
		case strings.TrimSpace(sec.Value) != "":
			tag = checkboxOn.Render("set")
		case sec.Required:
			tag = requiredStyle.Render("required, empty")
		default:
			tag = dimStyle.Render("empty")
		}
		item(k, tag)
	}

	section("Roles")
	if s.cfg.Cloud == stack.CloudGCP {
		g := s.cfg.GCP
		if g != nil {
			gcpOps := []struct {
				name       string
				perms      []string
				predefined string
			}{
				{"provision", g.ProvisionPermissions, g.ProvisionPredefinedRole},
				{"maintenance", g.MaintenancePermissions, g.MaintenancePredefinedRole},
				{"deprovision", g.DeprovisionPermissions, g.DeprovisionPredefinedRole},
			}
			for _, op := range gcpOps {
				item(op.name, roleStateTag(len(op.perms) > 0 || op.predefined != ""))
			}
			for _, k := range sortedKeys(g.BreakGlassRoles) {
				item(k+" "+dimStyle.Render("(break-glass)"), roleStateTag(g.BreakGlassRoles[k].Enabled))
			}
			for _, k := range sortedKeys(g.CustomRoles) {
				item(k+" "+dimStyle.Render("(custom)"), roleStateTag(g.CustomRoles[k].Enabled))
			}
		}
		return strings.Join(lines, "\n")
	}

	ops := []struct {
		name   string
		inline string
		perms  []string
		arns   []string
	}{
		{"provision", s.cfg.AWS.ProvisionInlinePolicyDocument, s.cfg.AWS.ProvisionPermissions, s.cfg.AWS.ProvisionManagedPolicyARNs},
		{"maintenance", s.cfg.AWS.MaintenanceInlinePolicyDocument, s.cfg.AWS.MaintenancePermissions, s.cfg.AWS.MaintenanceManagedPolicyARNs},
		{"deprovision", s.cfg.AWS.DeprovisionInlinePolicyDocument, s.cfg.AWS.DeprovisionPermissions, s.cfg.AWS.DeprovisionManagedPolicyARNs},
	}
	for _, op := range ops {
		on := hasOp(op.inline, op.perms, op.arns)
		item(op.name, roleStateTag(on))
	}
	for _, k := range sortedKeys(s.cfg.AWS.BreakGlassRoles) {
		item(k+" "+dimStyle.Render("(break-glass)"), roleStateTag(s.cfg.AWS.BreakGlassRoles[k].Enabled))
	}
	for _, k := range sortedKeys(s.cfg.AWS.CustomRoles) {
		item(k+" "+dimStyle.Render("(custom)"), roleStateTag(s.cfg.AWS.CustomRoles[k].Enabled))
	}

	return strings.Join(lines, "\n")
}

func roleStateTag(on bool) string {
	if on {
		return checkboxOn.Render("create")
	}
	return checkboxOff.Render("skip")
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

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
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
